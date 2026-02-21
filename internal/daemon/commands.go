package daemon

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
)

var (
	UnknownCommandType    = errors.New("Unknown command type")
	CommandNotExists      = errors.New("Command do not exists")
	IncorrectNumberOfArgs = errors.New("Incorrect number of arguments")
)

type commandFunc func(...string) ([]string, error)

type CommandWrap struct {
	minIn int
	maxIn int
	fn    commandFunc
}

func (command CommandWrap) MinIn() int {
	return command.minIn
}
func (command CommandWrap) MaxIn() int {
	return command.maxIn
}

func (command CommandWrap) ValidateArgCount(count int) error {
	if count >= command.minIn && count <= command.maxIn {
		return nil
	}

	if command.minIn == command.maxIn {
		return fmt.Errorf(
			"%w: got %v, expected %v",
			IncorrectNumberOfArgs,
			count, command.minIn)
	} else {
		return fmt.Errorf(
			"%w: got %v, expected between %v and %v",
			IncorrectNumberOfArgs,
			count, command.minIn, command.maxIn)
	}
}

func (command CommandWrap) Call(s ...string) ([]string, error) {
	if err := command.ValidateArgCount(len(s)); err != nil {
		return nil, err
	} else {
		return command.fn(s...)
	}
}

type CommandMap map[string]CommandWrap

type Commands struct {
	Actions CommandMap
	Setters CommandMap
	Queries CommandMap
	Fors    CommandMap
}

type CommandContext struct {
	TargetCid          ClientId
	TargetWokspaceNum  uint
	QueriedCid         ClientId
	QueriedWokspaceNum uint
}

var defaultCtx CommandContext
var ctx *CommandContext = &defaultCtx

func InitCommands(tracker *tracker) Commands {
	var workspaces map[uint]*Workspace

	if tracker != nil {
		workspaces = tracker.workspaces
		defaultCtx.TargetCid = ClientId(state.ActiveWin)
		defaultCtx.TargetWokspaceNum = state.CurrentDesk

	} else {
		// It's okay for it to be nil,
		// the actions will be discarded
		// at the end of this function
		workspaces = nil
	}

	keybindActions := map[string]func(){
		"tile": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.IsTiling = true
			ws.Tile()
		},
		"untile": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.Untile()
		},
		"make_active_window_master": func() {
			c := tracker.clients[state.ActiveWin]
			ws := workspaces[state.CurrentDesk]
			ws.ActiveLayout().MakeMaster(c)
			ws.Tile()
		},
		"switch_layout": func() {
			workspaces[ctx.TargetWokspaceNum].SwitchLayout()
		},
		"increase_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().IncMaster()
			ws.Tile()
		},
		"decrease_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().DecreaseMaster()
			ws.Tile()
		},
		"next_window": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().NextClient()
		},
		"previous_window": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().PreviousClient()
		},
		"increment_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().IncrementMaster()
			ws.Tile()
		},
		"decrement_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			ws.ActiveLayout().DecrementMaster()
			ws.Tile()
		},
	}

	actions := CommandMap{
		// Internal command, used for resetting context on each new command sequence
		"__start_new_command_sequence": CommandWrap{
			minIn: 0, maxIn: 0,
			fn: func(args ...string) ([]string, error) {
				ctx.TargetCid = ClientId(state.ActiveWin)
				ctx.TargetWokspaceNum = state.CurrentDesk
				return nil, nil
			},
		},
		"swap": CommandWrap{
			minIn: 1, maxIn: 2,
			fn: func(args ...string) ([]string, error) {
				var secondId, firstId ClientId
				var secondIdErr, firstIdErr error

				firstId, firstIdErr = ParseClientId(args[0])
				if firstIdErr != nil {
					firstIdErr = fmt.Errorf("Parse error for client id \"%v\": %w", args[0], firstIdErr)
				}

				if len(args) == 1 {
					secondId = ctx.TargetCid
				} else {
					secondId, secondIdErr = ParseClientId(args[1])
				}
				if secondIdErr != nil {
					secondIdErr = fmt.Errorf("Parse error for client id \"%v\": %w", args[1], secondIdErr)
				}

				if err := errors.Join(firstIdErr, secondIdErr); err != nil {
					return nil, err
				}

				ws := workspaces[ctx.TargetWokspaceNum]
				success := ws.ActiveLayout().SwapById(secondId, firstId)
				if !success {
					return nil, fmt.Errorf(
						"Cliend id %v was not found in current workspace",
						args[0],
					)
				}           
				
				ws.Tile()

				return nil, nil
			},
		},
	}

	// TODO: Remove when keybind dispatching will be redone
	for k, v := range keybindActions {
		actions[k] = CommandWrap{
			fn: wrapActionToCommandFunc(v),
		}
	}

	setters := CommandMap{
		"layout": CommandWrap{
			minIn: 1, maxIn: 1,
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
		"layout": CommandWrap{
			minIn: 0, maxIn: 0,
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
	fors := CommandMap{
		"window": CommandWrap{
			minIn: 1, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				cid, err := ParseClientId(args[0])
				if err != nil {
					err = fmt.Errorf("Parse error for client id \"%v\": %w", args[0], err)
				}

				ctx.TargetCid = cid

				return nil, err
			},
		},
		"workspace": CommandWrap{
			minIn: 1, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				workspaceNum, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					err = fmt.Errorf("Parse error for workspace number \"%v\": %w", args[0], err)
				} else if workspaceNum >= uint64(state.DeskCount) {
					err = fmt.Errorf("Parse error for workspace number \"%v\": number is out of range", args[0])
				}

				ctx.TargetWokspaceNum = uint(workspaceNum)

				return nil, err
			},
		},
	}

	return Commands{
		Actions: actions,
		Queries: queries,
		Setters: setters,
		Fors:    fors,
	}
}

// TODO: Remove when keybind dispatching will be redone
func wrapActionToCommandFunc(fn func()) commandFunc {
	return func(s ...string) ([]string, error) {
		fn()
		return nil, nil
	}
}

func (c Commands) Map(kind types.CommandType) CommandMap {
	switch kind {
	case types.Action:
		return c.Actions
	case types.Set:
		return c.Setters
	case types.Query:
		return c.Queries
	case types.For:
		return c.Fors
	default:
		return nil
	}
}

func (c Commands) GetByName(kind types.CommandType, name string) (CommandWrap, bool) {
	mapOfKind := c.Map(kind)
	if mapOfKind == nil {
		return CommandWrap{}, false
	}
	command, exists := mapOfKind[name]
	return command, exists
}

func (c Commands) Do(command types.Command) types.CommandResult {
	commandMap := c.Map(command.Kind)
	if commandMap == nil {
		return types.CommandResult{Messages: nil, Err: UnknownCommandType}
	}

	commandWrap, exists := commandMap[command.Name]
	if !exists {
		return types.CommandResult{Messages: nil, Err: CommandNotExists}
	} else {
		messages, err := commandWrap.Call(command.Args...)
		return types.CommandResult{Messages: messages, Err: err}
	}
}
