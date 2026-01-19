package main

import (
	"github.com/Alnivel/zentile/state"
)

type ActionFunc func()

func initActions(tracker *tracker) map[string]ActionFunc {
	var workspaces map[uint]*Workspace

	if tracker != nil {
		workspaces = tracker.workspaces
	} else {
		// It's okay for it to be nil,
		// the actions will be discarded
		// at the end of this function
		workspaces = nil
	}

	actions := map[string]ActionFunc{
		"tile": func() {
			ws := workspaces[state.CurrentDesk]
			ws.IsTiling = true
			ws.Tile()
		},
		"untile": func() {
			ws := workspaces[state.CurrentDesk]
			ws.Untile()
		},
		"make_active_window_master": func() {
			c := tracker.clients[state.ActiveWin]
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().MakeMaster(c)
			ws.Tile()
		},
		"switch_layout": func() {
			workspaces[state.CurrentDesk].SwitchLayout()
		},
		"increase_master": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().IncMaster()
			ws.Tile()
		},
		"decrease_master": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().DecreaseMaster()
			ws.Tile()
		},
		"next_window": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().NextClient()
		},
		"previous_window": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().PreviousClient()
		},
		"increment_master": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().IncrementMaster()
			ws.Tile()
		},
		"decrement_master": func() {
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().DecrementMaster()
			ws.Tile()
		},
	}

	if tracker == nil {
		for key := range actions {
			actions[key] = nil
		}
	}
	return actions
}
