#Horace: a GO IRC bot

I wanted a project that would help me learn [go](http://golang.com). I learned python with irc and [xmpp](http://wokkel.ik.nu/) [bots](http://excla.im/), i learned network programming in perl with irc bots - so it only seemed fair that I learn go by writing an IRC bot. 

I am using the event-based irc framework from: [fluffle/goirc](https://github.com/fluffle/goirc). It works very well and being event based, makes the code fun. 

##Installation

Install dependencies:

	go get github.com/gosexy/yaml
	go get github.com/fluffle/goirc/client
	go get github.com/gosexy/to
	go get github.com/fluffle/goevent/event
	go get github.com/fluffle/golog/logging
	go get github.com/gosexy/sugar
	go get github.com/gosexy/to
	go get github.com/harperreed/gobitly/bitly

Compile:
	
	go build bot.go

	
Run bot:

	./bot 

This will generate a blank config file.

	INFO Creating settings file: settings.yaml
	ERROR You must edit the settings.yaml file before continuing


Edit the config file `settings.yaml`:
	
	vi settings.yaml

And run the bot again

	./bot

Your bot will connect

	INFO Read configuration file: settings.yaml
	INFO create new IRC connection
	INFO irc.Connect(): Connecting to irc.example.net:6667 without SSL.
	INFO connected as gobo667
	INFO mode: "gobot" 
	INFO mode: "+i" 
	INFO Joined #example


##Config

Example `settings.yaml`. Rather self explanatory.


	bitly:
    	api_key: xxxxxxxxxxxxxxxxxxxxx
	    shorturls_enabled: true
	    username: example
  	bot_config:
    	channel_protection: true
	    friends:
	    - friend1!example@example/example
	    - friend2!example@example/example
	    - friend2!example@example/example
	    owner: example!example@example/example
    	rejoin_on_kick: true
    connection:
    	channel: '#example'
	    irc_server: irc.example.net
	    nick: gobot
	    realname: Go Bot

You can generate a new config by running the `./bot -config_file=settings.yaml -generate_config=true`

##Command line flags

	Usage of ./bot:
	  -channel="": IRC channel
	  -command_char="!": Command character
	  -config_file="settings.yaml": YAML config file
	  -generate_config=false: Generate a config file
	  -irc_server="": IRC server
	  -nick="": IRC nickname
	  -realname="Go Bot": IRC realname
	  -rejoin_on_kick=true: Rejoin on kick

##Todo

* Channel management
	* topic control
	* banning	
* enemies
* utilities (twitter, weather, etc)
* webhooks

I would also like to experiment with a way to make it a bit more plugable. 
