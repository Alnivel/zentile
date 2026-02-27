package backend

import (
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/jezek/xgbutil/xevent"
	log "github.com/sirupsen/logrus"
)

type X11Keybinder struct {
	X *xgbutil.XUtil
}

func newX11Keybinder(backend x11Backend) X11Keybinder {
	keybind.Initialize(backend.X)

	return X11Keybinder{backend.X}
}

func (k X11Keybinder) Bind(keyStr string, f func()) {
	if len(keyStr) == 0 {
		return
	}

	err := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			f()
		}).Connect(k.X, k.X.RootWin(),
		keyStr, true)

	if err != nil {
		log.Warn(err)
	}
}
