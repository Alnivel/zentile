package daemon

import (
	"sync"

	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/jezek/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

type MutexChan struct {
	Chan  chan<- types.Command
	Mutex sync.Mutex
}

type Keybindings struct {
	commands map[string][]types.Command
}

func (k Keybindings) HandleIncomingCommands(commandChan chan<- CommandRequest, chanMutex *sync.Mutex) {
	keybind.Initialize(state.X)

	for keyStr, command := range k.commands {
		bind(keyStr, func() {
			// Only one command sequence can be executed at the same time
			chanMutex.Lock()
			defer chanMutex.Unlock()
			for _, command := range command {
				commandRequest, replyChan := NewCommandRequest(command)

				commandChan <- commandRequest
				result := <-replyChan
				if result.Err != nil {
					log.Error(result.Err.Error())
					break
				}
			}
		})
	}
}

func bind(keyStr string, f func()) {
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
