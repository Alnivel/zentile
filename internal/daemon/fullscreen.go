package daemon

import (
	"github.com/Alnivel/zentile/internal/daemon/state"
	log "github.com/sirupsen/logrus"
)

type FullScreen struct {
	*Store
	WorkspaceNum uint
}

func (fs *FullScreen) Do() {
	log.Info("Switching to Fullscreen layout")
	for _, c := range fs.Store.All() {
		x, y, w, h := state.WorkAreaDimensions(fs.WorkspaceNum)
		c.MoveResize(x, y, w, h)
	}
}

func (fs *FullScreen) Undo() {
	for _, c := range append(fs.masters, fs.slaves...) {
		c.Restore()
	}
}

func (fs *FullScreen) GetProportion() float64 {
	return 1
}

func (fs *FullScreen) SetProportion(proportion float64) {
}

func (fs *FullScreen) sto() *Store {
	return fs.Store
}
