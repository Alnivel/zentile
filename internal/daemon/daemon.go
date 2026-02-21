package daemon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	commandparser "github.com/Alnivel/zentile/internal/command_parser"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
	"github.com/jezek/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

var Config config.Config

func Start(config config.Config, args []string) {
	pingQuit := make(chan struct{}, 1)
	go handleInterruptsGracefully(pingQuit)

	Config = config
	state.Populate()

	windowTracker := initTracker(CreateWorkspaces())
	commands := InitCommands(windowTracker)

	pingBeforeXEvent, pingAfterXEvent, pingXQuit := xevent.MainPing(state.X)
	commandChan := make(chan CommandRequest)
	commandChanMutex := sync.Mutex{}

	socketPath := "/tmp/zentile.sock"
	socketListener, err := ListenSocket(socketPath)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer socketListener.Close()
	socketListener.HandleIncomingCommands(commandChan, &commandChanMutex)

	getCommandByNameAdapter := func(kind types.CommandType, name string) (commandparser.CommandWrap, bool) {
		return commands.GetByName(kind, name)
	}
	commandParser := commandparser.CommandParser{GetCommandByName: getCommandByNameAdapter}

	commandKeybinings := make(map[string][]types.Command)
	for keyStr, commandStr := range config.Keybindings {
		commandKeybinings[keyStr], err = commandParser.ParseString(commandStr)
		if err != nil {
			log.Warn(err)
		}
	}

	keybindings := Keybindings{
		commandKeybinings,
	}
	keybindings.HandleIncomingCommands(commandChan, &commandChanMutex)

	for {
		select {
		case <-pingBeforeXEvent:
			// Wait for the event to finish processing.
			<-pingAfterXEvent

		case commandRequest := <-commandChan:
			commandRequest.SendResult(commands.Do(commandRequest.Command))

		case <-pingXQuit:
			return
		case <-pingQuit:
			return
		}
	}
}

func handleInterruptsGracefully(pingQuit chan<- struct{}) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	<-sigc
	log.Warn("Attempting to gracefully shutdown...")
	// pingQuit is buffered, so the send is not blocking
	pingQuit <- struct{}{}

	<-sigc
	log.Warn("Forcefully terminated")
	os.Exit(0)
}
