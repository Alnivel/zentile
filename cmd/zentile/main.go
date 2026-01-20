package main

import (
	"flag"
	"github.com/Alnivel/zentile/internal/daemon"
	"github.com/Alnivel/zentile/internal/cli"
	"github.com/Alnivel/zentile/internal/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	//TODO: Add separate parsing of args which applies on top of the config
	config, err := config.InitConfig()
	setLogLevel()

	runAsDaemon := true
	if runAsDaemon {
		daemon.Start(config)
	} else {
		// sending command
		cli.Run()
	}

}

func setLogLevel() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}
