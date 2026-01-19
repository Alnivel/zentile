package main

import (
	"flag"

	"github.com/Alnivel/zentile/state"
	"github.com/BurntSushi/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

func main() {
	//TODO: Add separate parsing of args which applies on top of the config
	setLogLevel()

	daemon := true
	if daemon {
		runDaemon()
	} else {
		// sending command
		runCli()
	}

}

func runCli() {
	socketChan := make(chan string)
	_ = socketChan

}

func runDaemon() {

	state.Populate()

	t := initTracker(CreateWorkspaces())
	actions := initActions(t)
	bindKeys(actions)

	//TODO: Implement opening and receiving from socket into channel
	pingBefore, pingAfter, pingQuit := xevent.MainPing(state.X)
	socketChan, err := ListenSocket()
	if err != nil {
		return
	}

	for {
		select {
		case <-pingBefore:
			// Wait for the event to finish processing.
			<-pingAfter
		case socketCommand := <-socketChan:
			action, exists := actions[socketCommand.name]
			if exists {
				action()
				socketCommand.ReplyOk()
			} else {
				// TODO: Add command name to the reply
				socketCommand.ReplyErr("Unknown command")
			}
		case <-pingQuit:
			return
		}
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
