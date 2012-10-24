package main

import (
	"flag"
	"fmt"
	"github.com/fluffle/goevent/event"
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/golog/logging"
	"github.com/gosexy/to"
	"github.com/gosexy/yaml"
	"strings"
)

// flags
var config_file *string = flag.String("config_file", "settings.yaml", "YAML config file")
var irc_server *string = flag.String("irc_server", "", "IRC server")
var channel *string = flag.String("channel", "", "IRC channel")
var nick *string = flag.String("nick", "", "IRC nickname")
var realname *string = flag.String("realname", "Go Bot", "IRC realname")
var rejoin_on_kick *bool = flag.Bool("rejoin_on_kick", true, "Rejoin on kick")
var command_char *string = flag.String("command_char", "!", "Command character")

type BotCommandHandler func(*irc.Conn, *irc.Line, []string)

func NewHandler(f BotCommandHandler) event.Handler {
	return event.NewHandler(func(ev ...interface{}) {
		f(ev[0].(*irc.Conn), ev[1].(*irc.Line), ev[2].([]string))
	})
}

func Trust(identity string, trusted_identities []string) bool {
	log := logging.InitFromFlags()
	for _, value := range trusted_identities {
		if identity == value {
			log.Info("Authenticated: " + identity)
			return true
		}
	}
	log.Info("Authentication failed for : " + identity)
	return false
}

func main() {

	// Parse flags from command line
	flag.Parse()
	log := logging.InitFromFlags()

	// setup logging
	log.SetLogLevel(2)

	// handle configuration

	log.Info("Read configuration file: " + *config_file)
	settings := yaml.New()
	settings.Read(*config_file)

	if *channel == "" {
		*channel = settings.Get("connection/channel").(string)
		log.Debug("Read channel from config file: " + *channel)
	} else {
		log.Debug("Read channel from flag: " + *channel)
	}

	if *nick == "" {
		*nick = settings.Get("connection/nick").(string)
		log.Debug("Read nick from config file: " + *nick)
	} else {
		log.Debug("Read nick from flag: " + *nick)
	}

	if *realname == "" {
		*realname = settings.Get("connection/realname").(string)
		log.Debug("Read realname from config file: " + *realname)
	} else {
		log.Debug("Read realname from flag: " + *realname)
	}

	if *irc_server == "" {
		*irc_server = settings.Get("connection/irc_server").(string)
		log.Debug("Read irc_server from config file: " + *irc_server)
	} else {
		log.Debug("Read irc_server from flag: " + *irc_server)
	}

	if *rejoin_on_kick == true {
		*rejoin_on_kick = settings.Get("bot_config/rejoin_on_kick").(bool)
		log.Debug("Read rejoin_on_kick from config file: %t ", *rejoin_on_kick)
	} else {
		log.Debug("Read rejoin_on_kick from flag: %t ", *rejoin_on_kick)
	}

	owner_nick := to.String(settings.Get("bot_config/owner"))
	friends := to.List(settings.Get("bot_config/friends"))
	trusted_identities := make([]string, 0)
	trusted_identities = append(trusted_identities, owner_nick)
	for _, value := range friends {
		trusted_identities = append(trusted_identities, value.(string))
	}

	// set up bot command event registry
	bot_command_registry := event.NewRegistry()
	reallyquit := false

	// Bot command handlers
	// addfriend
	addfriend_state := make(chan string)
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if line.Src == owner_nick {
			channel := line.Args[0]
			if len(commands) > 1 {
				target := commands[1]
				log.Debug("adding friend: %q", target)
				conn.Whois(target)
				log.Debug("triggering channel: %q", target)
				addfriend_state <- target

			} else {
				conn.Privmsg(channel, line.Nick+": use !addfriend <friend nick>")
			}
		}
	}), "addfriend")

	//save
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if line.Src == owner_nick {
			channel := line.Args[0]
			conn.Privmsg(channel, line.Nick+": saving settings")
			settings.Set("bot_config/friends", friends)
			settings.Write(*config_file)
			log.Info("%q", to.List(settings.Get("bot_config/friends")))
		}
	}), "save")

	//reload
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if line.Src == owner_nick {
			channel := line.Args[0]
			conn.Privmsg(channel, line.Nick+": reloading settings")
			friends := to.List(settings.Get("bot_config/friends"))
			trusted_identities = make([]string, 0)
			trusted_identities = append(trusted_identities, owner_nick)
			for _, value := range friends {
				trusted_identities = append(trusted_identities, value.(string))
			}
			log.Info("%q", to.List(settings.Get("bot_config/friends")))
		}
	}), "reload")

	// op
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if Trust(line.Src, trusted_identities) {
			channel := line.Args[0]
			if len(commands) > 1 {
				target := commands[1]
				log.Info("Oping user: " + target)
				conn.Mode(channel, "+o "+target)
			} else {
				log.Info("Oping user: " + line.Nick)
				conn.Mode(channel, "+o "+line.Nick)
			}
		}
	}), "op")

	//deop
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if Trust(line.Src, trusted_identities) {
			channel := line.Args[0]
			if len(commands) > 1 {
				target := commands[1]
				conn.Mode(channel, "-o "+target)
			} else {
				conn.Mode(channel, "-o "+line.Nick)
			}
		}
	}), "deop")

	// kick
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if Trust(line.Src, trusted_identities) {
			channel := line.Args[0]
			if len(commands) > 1 {
				target := commands[1]
				kick_message := "get out"
				if len(commands) > 2 {
					//this doesn't work. Need to fix. 
					kick_message = commands[2]
				}
				//do'nt kick if owner
				if target != strings.Split(owner_nick, "!")[0] {
					//don't kick if self
					if *nick != target {
						conn.Kick(channel, target, kick_message)
					} else {
						conn.Privmsg(channel, line.Nick+": why would i kick myself?")
					}
				} else {
					conn.Privmsg(channel, line.Nick+": why would i kick my lovely friend "+target+"?")

				}
			} else {
				conn.Privmsg(channel, line.Nick+": invalid command")
			}
		}
	}), "kick")

	//quit
	bot_command_registry.AddHandler(NewHandler(func(conn *irc.Conn, line *irc.Line, commands []string) {
		if line.Src == owner_nick {
			quit_message := "i died"
			reallyquit = true
			if len(commands) > 1 {
				quit_message = commands[1]
			}
			conn.Quit(quit_message)
		}
	}), "quit")

	// create new IRC connection
	log.Info("create new IRC connection")
	irc_client := irc.SimpleClient(*nick, *realname)

	// IRC HANDLERS!

	irc_client.EnableStateTracking()
	irc_client.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("connected as " + *nick)
			conn.Join(*channel)
		})

	// Set up a handler to notify of disconnect events.
	quit := make(chan bool)
	irc_client.AddHandler("disconnected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("disconnected")
			quit <- true
		})

	//Handle Private messages
	irc_client.AddHandler("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("privmsg")
			irc_input := strings.ToLower(line.Args[1])
			if strings.HasPrefix(irc_input, *command_char) {
				irc_command := strings.Split(irc_input[1:], " ")
				bot_command_registry.Dispatch(irc_command[0], conn, line, irc_command)
			}

		})

	//handle kick by rejoining kicked channel
	irc_client.AddHandler("KICK",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Kicked from " + line.Args[0])
			if *rejoin_on_kick {
				log.Info("rejoining " + line.Args[0])
				conn.Join(line.Args[0])
			}
		})

	//notify on MODE
	irc_client.AddHandler("MODE",
		func(conn *irc.Conn, line *irc.Line) {
			for _, v := range line.Args {
				log.Info("mode: %q ", v)
			}
		})

	//notify on MODE
	irc_client.AddHandler("311",
		func(conn *irc.Conn, line *irc.Line) {
			addedfriend := <-addfriend_state
			log.Info("addfriend channel: %q", addedfriend)
			if addedfriend == line.Args[1] {
				friend := line.Args[1] + "!" + line.Args[2] + "@" + line.Args[3]
				log.Debug("added friend " + friend)
				trusted_identities = append(trusted_identities, friend)
				friends = append(friends, friend)
				log.Debug("friends: %q", friends)
				conn.Privmsg(strings.Split(owner_nick, "!")[0], line.Nick+": added "+line.Args[1]+" as friend")
				//addfriend_state <- ""
			} else {
				log.Info("addfriend channel is empty: %q", addedfriend)

			}
		})

	//notify on join
	irc_client.AddHandler("JOIN",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Joined " + line.Args[0])
		})

	//handle topic changes 
	irc_client.AddHandler("TOPIC",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Topic on " + line.Args[0] + " changed to: " + line.Args[1])
		})

	// set up a goroutine to read commands from stdin

	for !reallyquit {
		// connect to server
		if err := irc_client.Connect(*irc_server); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}

		// wait on quit channel
		<-quit
	}
}
