package daemon

import (
	"math"

	"github.com/Alnivel/zentile/internal/config"
)

const (
	MASTER_MAX_PROPORTION = 0.9
	MASTER_MIN_PROPORTION = 0.1
)

type Layout interface {
	Do()
	Undo()
	Add(client Client)
	Remove(client Client)

	MakeMaster(client Client) bool
	Swap(this Client, that Client) bool

	IncMaster()
	DecreaseMaster()
	ClientRelative(relativeTo Client, offset int) (Client, bool)

	GetProportion() float64
	SetProportion(proportion float64)
	sto() *Store
}

type VertHorz struct {
	*Store
	Proportion   float64
	WorkspaceNum uint
	Tracker Tracker
	Config *config.WorkspaceConfig
}

func (l *VertHorz) Undo() {
	for _, c := range append(l.masters, l.slaves...) {
		c.Restore()
	}
}

func (l *VertHorz) GetProportion() float64 {
	return l.Proportion
}

func (l *VertHorz) SetProportion(proportion float64) {
	l.Proportion = math.Min(math.Max(proportion, MASTER_MIN_PROPORTION), MASTER_MAX_PROPORTION)
}

func (l *VertHorz) sto() *Store {
	return l.Store
}
