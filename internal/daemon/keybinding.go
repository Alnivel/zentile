package daemon

import (
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

type keyMapper struct{}

func (k keyMapper) bind(action string, f func()) {
	keyStr := Config.Keybindings[action]
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

func bindKeys(actions map[string]ActionFunc) {
	keybind.Initialize(state.X)
	k := keyMapper{}

	for actionname, actionfunc := range actions {
		k.bind(actionname, actionfunc)
	}
}
