package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/daemon"
)

const defaultConfig = "~/.config/tv2go/config.json"

func loadConfig(configFile string) *config.Config {
	if len(configFile) == 0 {
		glog.Infof("No --config_file given.  Using default: %s", defaultConfig)
		configFile = defaultConfig
	}

	glog.Info("Got config file: %s", configFile)
	config := config.NewConfig()
	err := config.ReadConfig(configFile)
	if err != nil {
		glog.Fatal(err.Error())
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
