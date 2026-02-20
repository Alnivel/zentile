package daemon

import (
	"sync"

	commandparser "github.com/Alnivel/zentile/internal/command_parser"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/jezek/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

type keyMapper struct{}

func (k keyMapper) bind(keyStr string, f func()) {
	if len(keyStr) == 0 {
		return
	}

	err := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			f()
		}).Connect(state.X, state.X.RootWin(),
		keyStr, true)

	if err != nil {
		log.Warn(err)
	}
}

func HandleKeybindings(config config.Config, commandParser commandparser.CommandParser) (<-chan types.Command, chan<- struct{}) {
	keybind.Initialize(state.X)
	k := keyMapper{}

	commandChan := make(chan types.Command)
	commandDonePing := make(chan struct{})
	mutex := sync.Mutex{} // Only one command can be executed at the same time

	for keybinding, command := range config.Keybindings {
		parsedCommands, err := commandParser.ParseString(command)
		if err != nil {
			log.Warn(err)
		}

		k.bind(keybinding, func() {
			for _, command := range parsedCommands {
				mutex.Lock()
				commandChan <- command
				<-commandDonePing
				mutex.Unlock()
			}
		})
	}

	return commandChan, commandDonePing
}
