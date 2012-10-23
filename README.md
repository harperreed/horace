#horace: go IRC bot

I wanted a project that would help me learn go. I learned python with irc and xmpp bots, i learned network programming in perl with irc bots - so it only seemed fair that I learn go by writing an IRC bot. 

I am using the event-based irc framework from: [fluffle/goirc](https://github.com/fluffle/goirc). It works very well and being event based, makes the code fun. 

##Installation

Install dependencies:

  	go get github.com/gosexy/yaml
	go get github.com/fluffle/goirc/client
	go get github.com/gosexy/to

Compile:
	
	go build build_settings_yaml.go
	go build bot.go

Generate and edit the config file `settings.yaml`:
	
	./build_settings_yaml
	vi settings.yaml
	
Run bot:

	./bot 
	

##Config

Example `config.yaml`. Rather self explanatory.

	bot_config:
	  rejoin_on_kick: true
	connection:
	  channel: '#example'
	  irc_server: irc.example.net
	  nick: gobot
	  realname: Go Bot

Generate by running the `build_settings_yaml`

##Todo

Currently the bot idles and returns to a channel on being kicked. 

I would like it to do at a minimum the following:

* short urling
* Channel management
	* oping (done)
	* topic control
	* kicking (doneish)
	* banning	
* friends/enemies (friends doneish)
* utilities (twitter, weather, etc)
* new name (done)
* webhooks

I would also like to experiment with a way to make it a bit more plugable. 
