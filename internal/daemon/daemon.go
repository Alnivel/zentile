package daemon

import (
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/BurntSushi/xgbutil/xevent"
)

var Config config.Config

func Start(config config.Config) {
	Config = config
	state.Populate()

	t := initTracker(CreateWorkspaces())
	actions := initActions(t)
	bindKeys(actions)

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
