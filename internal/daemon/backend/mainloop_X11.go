package backend

import "github.com/jezek/xgbutil/xevent"

func newX11MainLoop(backend x11Backend) (chan struct{}, chan struct{}, chan struct{}) {
	return xevent.MainPing(backend.X) 
}
