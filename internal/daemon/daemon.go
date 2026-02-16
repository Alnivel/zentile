package daemon

import (
	"os"
	"os/signal"
	"syscall"

	commandparser "github.com/Alnivel/zentile/internal/command_parser"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
	"github.com/BurntSushi/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

var Config config.Config

func Start(config config.Config, args []string) {
	Config = config
	state.Populate()

	t := initTracker(CreateWorkspaces())
	commands := InitCommands(t)

	pingBeforeXEvent, pingAfterXEvent, pingXQuit := xevent.MainPing(state.X)

	socketPath := "/tmp/zentile.sock"
	socketListener, err := ListenSocket(socketPath)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer socketListener.Close()

	pingQuit := make(chan struct{}, 1)
	go handleInterruptsGracefully(pingQuit)

	getCommandByNameAdapter := func(kind types.CommandType, name string) (commandparser.CommandWrap, bool) {
		return commands.GetByName(kind, name)
	}
	commandParser := commandparser.CommandParser{GetCommandByName: getCommandByNameAdapter}
	keybingingCommandChan, keybindingCommandDonePing := HandleKeybindings(config, commandParser)
	socketCommandChan, socketResultChan := socketListener.HandleIncomingCommands()

	for {
		select {
		case <-pingBeforeXEvent:
			// Wait for the event to finish processing.
			<-pingAfterXEvent

		case command := <-socketCommandChan:
			socketResultChan <- commands.Do(command)

		case command := <-keybingingCommandChan:
			result := commands.Do(command)
			if result.Err != nil {
				log.Error(result.Err.Error())
			}
			keybindingCommandDonePing <- struct{}{}

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
