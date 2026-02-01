package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/BurntSushi/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

var Config config.Config

func Start(config config.Config, args []string) {
	Config = config
	state.Populate()

	t := initTracker(CreateWorkspaces())
	commands := InitCommands(t)
	bindKeys(commands.KeybindActions)

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

	pingBeforeCommand, pingAfterCommand := socketListener.HandleIncomingCommands(commands)

	for {
		select {
		case <-pingBeforeXEvent:
			// Wait for the event to finish processing.
			<-pingAfterXEvent
		case <-pingBeforeCommand:
			// Wait for the event to finish processing.
			<-pingAfterCommand
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
