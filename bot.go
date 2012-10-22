package main

import (
	"bufio"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/fluffle/golog/logging"
	"github.com/gosexy/yaml"
	"os"
	"strings"
)

// flags
var config_file *string = flag.String("config_file", "settings.yaml", "YAML config file")
var irc_server *string = flag.String("irc_server", "", "IRC server")
var channel *string = flag.String("channel", "", "IRC channel")
var nick *string = flag.String("nick", "", "IRC nickname")
var realname *string = flag.String("realname", "Go Bot", "IRC realname")
var rejoin_on_kick *bool = flag.Bool("rejoin_on_kick", true, "Rejoin on kick")

func main() {

	// Parse flags from command line
	flag.Parse()
	log := logging.InitFromFlags()

	// setup logging

	//log.SetLogLevel(2)

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

	// create new IRC connection
	log.Info("create new IRC connection")
	c := irc.SimpleClient(*nick, *realname)

	// HANDLERS!

	c.EnableStateTracking()
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("connected as " + *nick)
			conn.Join(*channel)
		})

	// Set up a handler to notify of disconnect events.
	quit := make(chan bool)
	c.AddHandler("disconnected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("disconnected")
			quit <- true
		})

	//Handle Private messages
	c.AddHandler("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("privmsg")
			log.Info(line.Raw)
			log.Info(line.Cmd)
			log.Info(line.Args[0])
			log.Info("Message received on " + line.Args[0] + " from " + line.Nick + ": " + line.Args[1])
			c.Privmsg(line.Args[0], "ECHO: "+line.Args[1])
		})

	//handle kick by rejoining kicked channel
	c.AddHandler("KICK",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Kicked from " + line.Args[0])
			if *rejoin_on_kick {
				log.Info("rejoining " + line.Args[0])
				conn.Join(line.Args[0])
			}
		})

	//notify on join
	c.AddHandler("JOIN",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Joined " + line.Args[0])
		})

	//handle topic changes 
	c.AddHandler("TOPIC",
		func(conn *irc.Conn, line *irc.Line) {
			log.Info("Topic on " + line.Args[0] + " changed to: " + line.Args[1])
		})

	//GO ROUTINES!

	// set up a goroutine to read commands from stdin
	in := make(chan string, 4)
	reallyquit := false
	go func() {
		con := bufio.NewReader(os.Stdin)
		for {
			s, err := con.ReadString('\n')
			if err != nil {
				// wha?, maybe ctrl-D...
				close(in)
				break
			}
			// no point in sending empty lines down the channel
			if len(s) > 2 {
				in <- s[0 : len(s)-1]
			}
		}
	}()

	// set up a goroutine to do parsey things with the stuff from stdin
	go func() {
		for cmd := range in {
			if cmd[0] == ':' {
				switch idx := strings.Index(cmd, " "); {
				case cmd[1] == 'd':
					fmt.Printf(c.String())
				case cmd[1] == 'f':
					if len(cmd) > 2 && cmd[2] == 'e' {
						// enable flooding
						c.Flood = true
					} else if len(cmd) > 2 && cmd[2] == 'd' {
						// disable flooding
						c.Flood = false
					}
					for i := 0; i < 20; i++ {
						c.Privmsg("#", "flood test!")
					}
				case idx == -1:
					continue
				case cmd[1] == 'x':
					c.Privmsg("#", "test!")
				case cmd[1] == 'q':
					reallyquit = true
					c.Quit(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'j':
					c.Join(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'p':
					c.Part(cmd[idx+1 : len(cmd)])
				}
			} else {
				c.Raw(cmd)
			}
		}
	}()

	for !reallyquit {
		// connect to server
		if err := c.Connect(*irc_server); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}

		// wait on quit channel
		<-quit
	}
}
