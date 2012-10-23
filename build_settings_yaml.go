package main

import (
	"fmt"
	"github.com/gosexy/sugar"
	"github.com/gosexy/yaml"
	"os"
)

func main() {

	settings_filename := "settingss.yaml"

	if _, err := os.Stat(settings_filename); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Creating settings file: " + settings_filename)
			settings := yaml.New()
			settings.Set("connection/irc_server", "irc.example.net")
			settings.Set("connection/channel", "#example")
			settings.Set("connection/nick", "gobot")
			settings.Set("connection/realname", "Go Bot")

			settings.Set("bot_config/rejoin_on_kick", true)
			settings.Set("bot_config/owner", "example!example@example/example")
			settings.Set("bot_config/friends", sugar.List{"friend1!example@example/example", "friend2!example@example/example", "friend2!example@example/example"})
			settings.Write(settings_filename)
		} else {

		}
	} else {
		fmt.Println("Settings file: " + settings_filename + " already exists")
	}

}
