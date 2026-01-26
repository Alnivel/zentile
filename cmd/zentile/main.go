package main

import (
	"flag"
	"github.com/Alnivel/zentile/internal/cli"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon"
	log "github.com/sirupsen/logrus"
)

type Flags struct {
	verbose bool
}

type Args []string

func parseArgs() (Args, Flags) {
	flags := Flags{}
	flag.BoolVar(&flags.verbose, "v", false, "verbose mode")
	flag.Parse()

	return flag.Args(), flags
}

func main() {
	args, flags := parseArgs()

	config, err := config.InitConfig()
	if err != nil {
		return
	}
	setLogLevel(flags.verbose)

	runAsDaemon := len(args) == 0
	if runAsDaemon {
		daemon.Start(config, args)
	} else {
		// sending command
		cli.Run(config, args)
	}

}

func setLogLevel(verbose bool) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}
