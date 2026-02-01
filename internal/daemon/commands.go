package daemon

import (
	"errors"
	"fmt"

	"github.com/Alnivel/zentile/internal/daemon/state"
)

var (
	UnknownCommandType    = errors.New("Unknown command type")
	CommandNotExists      = errors.New("Command do not exists")
	IncorrectNumberOfArgs = errors.New("Incorrect number of arguments")
)

type commandFunc func(...string) ([]string, error)

type Command struct {
	MinIn int
	MaxIn int
	fn    commandFunc
}

func (command Command) validateArgCount(count int) error {
	if count >= command.MinIn && count <= command.MaxIn {
		return nil
	}

	if command.MinIn == command.MaxIn {
		return fmt.Errorf(
			"%w: got %v, expected %v",
			IncorrectNumberOfArgs,
			count, command.MinIn)
	} else {
		return fmt.Errorf(
			"%w: got %v, expected between %v and %v",
			IncorrectNumberOfArgs,
			count, command.MinIn, command.MaxIn)
	}
}

func (command Command) do(s ...string) ([]string, error) {
	if err := command.validateArgCount(len(s)); err != nil {
		return nil, err
	} else {
		return command.fn(s...)
	}
}

type ActionFunc func()
type CommandMap map[string]Command

type Commands struct {
	Actions CommandMap
	Setters CommandMap
	Queries CommandMap
	// Temporary field, until dispatching from keybinds will be redone
	KeybindActions map[string]ActionFunc
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

	keybindActions := map[string]ActionFunc{
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

	actions := CommandMap{
		"swap": Command{
			MinIn: 2, MaxIn: 2,
			fn: func (args ...string) ([]string, error) {
				return []string{"swop", args[0], args[1]}, nil
			},
		},
	}

	// TODO: Remove when keybind dispatching will be redone
	for k, v := range keybindActions {
		actions[k] = Command{
			fn: wrapActionToCommandFunc(v),
		}
	}

	setters := CommandMap{
		"layout": Command{
			MinIn: 1, MaxIn: 1,
			fn: func(args ...string) ([]string, error) {
				layoutName := args[0]
				ws := workspaces[state.CurrentDesk]

				if layoutName == "none" {
					ws.Untile()
					return nil, nil
				} else {
					return nil, ws.SetLayoutByName(layoutName)
				}
			},
		},
	}
	queries := CommandMap{
		"layout": Command{
			MinIn: 0, MaxIn: 0,
			fn: func(args ...string) ([]string, error) {
				ws := workspaces[state.CurrentDesk]
				if ws.IsTiling {
					return []string{ws.ActiveLayoutName()}, nil
				} else {
					return []string{"none"}, nil
				}
			},
		},
	}

	return Commands{
		Actions:        actions,
		Queries:        queries,
		Setters:        setters,
		KeybindActions: keybindActions,
	}
}

// TODO: Remove when keybind dispatching will be redone
func wrapActionToCommandFunc(fn ActionFunc) commandFunc {
	return func(s ...string) ([]string, error) {
		fn()
		return nil, nil
	}
}

type CommandType string

const (
	Action CommandType = "ACTION"
	Set    CommandType = "SET"
	Query  CommandType = "QUERY"
)

func (c Commands) Map(kind CommandType) CommandMap {
	switch kind {
	case Action:
		return c.Actions
	case Set:
		return c.Setters
	case Query:
		return c.Queries
	default:
		return nil
	}
}

func (c Commands) Do(kind CommandType, name string, args ...string) ([]string, error) {
	commandMap := c.Map(kind)
	if commandMap == nil {
		return nil, UnknownCommandType
	}

	command, exists := commandMap[name]
	if !exists {
		return nil, CommandNotExists
	} else {
		return command.do(args...)
	}
}
