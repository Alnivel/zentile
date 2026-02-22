package daemon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon/state"
	"github.com/Alnivel/zentile/internal/types"
)

var (
	UnknownCommandType    = errors.New("Unknown command type")
	CommandNotExists      = errors.New("Command do not exists")
	IncorrectNumberOfArgs = errors.New("Incorrect number of arguments")
	NoWindowInWorkspace   = errors.New("No target window found in target workspace")
)

type CommandMap map[string]CommandWrap

type Commands struct {
	Actions CommandMap
	Setters CommandMap
	Queries CommandMap
	Fors    CommandMap
}

type CommandContext struct {
	TargetCid         ClientId
	TargetWokspaceNum uint
	CidVariables      map[string]ClientId
}

var defaultCtx CommandContext

func InitCommands(tracker *tracker, config *config.Config) Commands {
	var ctx *CommandContext = &defaultCtx

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
		"increment_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			layout := ws.ActiveLayout()
			layout.SetProportion(layout.GetProportion() + config.Proportion)
			ws.Tile()
		},
		"decrement_master": func() {
			ws := workspaces[ctx.TargetWokspaceNum]
			layout := ws.ActiveLayout()
			layout.SetProportion(layout.GetProportion() - config.Proportion)
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
		"next_window": CommandWrap{
			minIn: 0, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				offset := 1

				if len(args) == 1 {
					parsedOffset, err := strconv.ParseInt(args[0], 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Parse error for offset \"%v\": %w", args[0], err)
					}
					offset = int(parsedOffset)
				}

				ws := workspaces[ctx.TargetWokspaceNum]
				prevWindow, found := ws.ActiveLayout().ClientRelative(ctx.TargetCid, offset)
				if !found {
					return nil, NoWindowInWorkspace
				}
				prevWindow.Activate()
				return nil, nil
			},
		},
		"previous_window": CommandWrap{
			minIn: 0, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				offset := 1

				if len(args) == 1 {
					parsedOffset, err := strconv.ParseInt(args[0], 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Parse error for offset \"%v\": %w", args[0], err)
					}
					offset = int(parsedOffset)
				}

				ws := workspaces[ctx.TargetWokspaceNum]
				prevWindow, found := ws.ActiveLayout().ClientRelative(ctx.TargetCid, -offset)
				if !found {
					return nil, NoWindowInWorkspace
				}
				prevWindow.Activate()
				return nil, nil
			},
		},
		"swap": CommandWrap{
			minIn: 1, maxIn: 2,
			fn: func(args ...string) ([]string, error) {
				var secondId, firstId ClientId
				var secondIdErr, firstIdErr error

				firstId, firstIdErr = parseClientId(args[0], ctx)

				if len(args) == 1 {
					secondId = ctx.TargetCid
				} else {
					secondId, secondIdErr = parseClientId(args[1], ctx)
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
				ws := workspaces[ctx.TargetWokspaceNum]

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
				ws := workspaces[ctx.TargetWokspaceNum]
				if ws.IsTiling {
					return []string{ws.ActiveLayoutName()}, nil
				} else {
					return []string{"none"}, nil
				}
			},
		},
		"next_window": CommandWrap{
			minIn: 0, maxIn: 1,
			fn: func(args ...string) ([]string, error) {

				var offset int64 = 1
				if len(args) == 1 {
					var err error
					offset, err = strconv.ParseInt(args[0], 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Parse error for offset \"%v\": %w", args[0], err)
					}
				}

				ws := workspaces[ctx.TargetWokspaceNum]
				window, found := ws.ActiveLayout().ClientRelative(ctx.TargetCid, int(offset))
				if !found {
					return nil, NoWindowInWorkspace
				}
				ctx.CidVariables["%queried"] = window.Id

				return []string{window.Id.String()}, nil
			},
		},
	}
	fors := CommandMap{
		"window": CommandWrap{
			minIn: 1, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				cid, err := parseClientId(args[0], ctx)
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

	commandCollection := Commands{
		Actions: actions,
		Queries: queries,
		Setters: setters,
		Fors:    fors,
	}

	return commandCollection
}

func parseClientId(arg string, ctx *CommandContext) (ClientId, error) {
	var id ClientId
	var err error = nil

	switch {
	case arg == "%target":
		id = ctx.TargetCid
	case arg == "%queried":
		fallthrough
	case strings.HasPrefix(arg, "%"):
		var exists bool
		id, exists = ctx.CidVariables[arg]
		if !exists {
			err = fmt.Errorf("Parse error for client id \"%v\": variable do not exists", arg)
		}
	default:
		id, err := ParseClientId(arg)
		if err != nil {
			return id, fmt.Errorf("Parse error for client id \"%v\": %w", arg, err)
		}
	}

	return id, err
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
