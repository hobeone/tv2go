package main

import (
	"flag"

	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/daemon"
	"github.com/mgutz/logxi/v1"
)

const defaultConfig = "~/.config/tv2go/config.json"

func loadConfig(configFile string) *config.Config {
	if len(configFile) == 0 {
		log.Info("No --config_file given.  Using default.", "file", defaultConfig)
		configFile = defaultConfig
	}

	log.Info("Got config file:", "file", configFile)
	config := config.NewConfig()
	err := config.ReadConfig(configFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	return config
}

func main() {
	cfgfile := flag.String("config_file", defaultConfig, "Config file to use")

	flag.Parse()

	cfg := loadConfig(*cfgfile)

	d := daemon.NewDaemon(cfg)
	d.Run()
}
