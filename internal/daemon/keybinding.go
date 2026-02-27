package daemon

import (
	"sync"

	"github.com/Alnivel/zentile/internal/types"
	log "github.com/sirupsen/logrus"
)

type MutexChan struct {
	Chan  chan<- types.Command
	Mutex sync.Mutex
}

type Keybindings struct {
	keybinder Keybinder
	commands map[string][]types.Command
}

func (k Keybindings) HandleIncomingCommands(commandChan chan<- CommandRequest, chanMutex *sync.Mutex) {
	for keyStr, command := range k.commands {
		k.keybinder.Bind(keyStr, func() {
			chanMutex.Lock()
			defer chanMutex.Unlock()
			requestStartNewCommandSequence(commandChan)

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

