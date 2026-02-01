package daemon

import (
	"errors"
	"fmt"

	"github.com/Alnivel/zentile/internal/daemon/state"
)

type ActionFunc func()
type SetFunc func(...string) error
type QueryFunc func(...string) ([]string, error)

type Commands struct {
	Actions map[string]ActionFunc
	Setters map[string]SetFunc
	Queries map[string]QueryFunc
}

func InitCommands(tracker *tracker) Commands {
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
	setters := map[string]SetFunc{
		"layout": func(args ...string) error {
			if err := CheckArgsCount(args, 1, 1); err != nil {
				return err
			}

			layoutName := args[0]
			ws := workspaces[state.CurrentDesk]

			if layoutName == "none" {
				ws.Untile()
				return nil
			} else {
				return ws.SetLayoutByName(layoutName)
			}
		},
	}
	queries := map[string]QueryFunc{
		"layout": func(args ...string) ([]string, error) {
			if err := CheckArgsCount(args, 0, 0); err != nil {
				return nil, err
			}
			ws := workspaces[state.CurrentDesk]
			if ws.IsTiling {
				return []string{ws.ActiveLayoutName()}, nil
			} else {
				return []string{"none"}, nil
			}
		},
	}

	if tracker == nil {
		nilMapValues(actions)
		nilMapValues(queries)
		nilMapValues(setters)
	}
	return Commands{
		Actions: actions,
		Queries: queries,
		Setters: setters,
	}
}

func nilMapValues[K comparable, V any](map_ map[K]V) {
	var nilAtHome V
	for key := range map_ {
		map_[key] = nilAtHome
	}
}

func CheckArgsCount(args []string, min int, max int) error {
	count := len(args)
	if count >= min && count <= max {
		return nil
	}

	if min == max {
		return fmt.Errorf(
			"%w: got %v, expected %v",
			IncorrectNumberOfArgs,
			count, min)
	} else {
		return fmt.Errorf(
			"%w: got %v, expected between %v and %v",
			IncorrectNumberOfArgs,
			count, min, max)
	}
}

var (
	UnknownCommandType    = errors.New("Unknown command type")
	CommandNotExists      = errors.New("Command do not exists")
	IncorrectNumberOfArgs = errors.New("Incorrect number of arguments")
)

type CommandType string

const (
	Action = "ACTION"
	Set    = "SET"
	Query  = "QUERY"
)

func (c Commands) Do(kind CommandType, name string, args ...string) ([]string, error) {
	switch kind {
	case Action:
		action, exists := c.Actions[name]
		if exists {
			action()
			return nil, nil
		} else {
			return nil, CommandNotExists
		}
	case Set:
		setter, exists := c.Setters[name]
		if exists {
			return nil, setter(args...)
		} else {
			return nil, CommandNotExists
		}
	case Query:
		query, exists := c.Queries[name]
		if exists {
			return query(args...)
		} else {
			return nil, CommandNotExists
		}
	default:
		return nil, UnknownCommandType
	}
}
