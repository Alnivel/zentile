package daemon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Alnivel/zentile/internal/config"
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
	TargetClient       Client
	TargetWorkspaceNum uint
	Variables          map[string]Client
}

var defaultCtx CommandContext

func InitCommands(tracker Tracker, config *config.Config) Commands {
	var ctx *CommandContext = &defaultCtx

	if tracker != nil {
		if client, exists := tracker.ActiveClient(); exists {
			defaultCtx.TargetClient = client
		}
		defaultCtx.TargetWorkspaceNum = tracker.CurentWorkspaceNum()
	} else {
		// It's okay for it to be nil,
		// the actions will be discarded
		// at the end of this function
	}

	keybindActions := map[string]func(){
		"tile": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			ws.isTiling = true
			ws.Tile()
		},
		"untile": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			ws.Untile()
		},
		"make_active_window_master": func() {
			ws := tracker.ActiveWorkspace()
			client, exists := tracker.ActiveClient()
			if exists {
				ws.ActiveLayout().MakeMaster(client)
				ws.Tile()
			}
		},
		"switch_layout": func() {
			tracker.Workspace(ctx.TargetWorkspaceNum).SwitchLayout()
		},
		"increase_master": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			ws.ActiveLayout().IncMaster()
			ws.Tile()
		},
		"decrease_master": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			ws.ActiveLayout().DecreaseMaster()
			ws.Tile()
		},
		"increment_master": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			layout := ws.ActiveLayout()
			layout.SetProportion(layout.GetProportion() + config.ProportionStep)
			ws.Tile()
		},
		"decrement_master": func() {
			ws := tracker.Workspace(ctx.TargetWorkspaceNum)
			layout := ws.ActiveLayout()
			layout.SetProportion(layout.GetProportion() - config.ProportionStep)
			ws.Tile()
		},
	}

	actions := CommandMap{
		// Internal command, used for resetting context on each new command sequence
		"__start_new_command_sequence": CommandWrap{
			minIn: 0, maxIn: 0,
			fn: func(args ...string) ([]string, error) {
				if client, exists := tracker.ActiveClient(); exists {
					ctx.TargetClient = client
				}
				ctx.TargetWorkspaceNum = tracker.CurentWorkspaceNum()
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

				ws := tracker.Workspace(ctx.TargetWorkspaceNum)
				nextClient, found := ws.ActiveLayout().ClientRelative(ctx.TargetClient, offset)
				if !found {
					return nil, NoWindowInWorkspace
				}
				nextClient.Activate()
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

				ws := tracker.Workspace(ctx.TargetWorkspaceNum)
				prevClient, found := ws.ActiveLayout().ClientRelative(ctx.TargetClient, -offset)
				if !found {
					return nil, NoWindowInWorkspace
				}
				prevClient.Activate()
				return nil, nil
			},
		},
		"swap": CommandWrap{
			minIn: 1, maxIn: 2,
			fn: func(args ...string) ([]string, error) {
				var secondClient, firstClient Client
				var secondIdErr, firstIdErr error

				firstClient, firstIdErr = parseClient(args[0], ctx, tracker)

				if len(args) == 1 {
					secondClient = ctx.TargetClient
				} else {
					secondClient, secondIdErr = parseClient(args[1], ctx, tracker)
				}

				if err := errors.Join(firstIdErr, secondIdErr); err != nil {
					return nil, err
				}

				ws := tracker.Workspace(ctx.TargetWorkspaceNum)
				success := ws.ActiveLayout().Swap(secondClient, firstClient)
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
				ws := tracker.Workspace(ctx.TargetWorkspaceNum)

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
				ws := tracker.Workspace(ctx.TargetWorkspaceNum)
				if ws.isTiling {
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

				ws := tracker.Workspace(ctx.TargetWorkspaceNum)
				client, found := ws.ActiveLayout().ClientRelative(ctx.TargetClient, int(offset))
				if !found {
					return nil, NoWindowInWorkspace
				}
				ctx.Variables["%queried"] = client

				return []string{client.String()}, nil
			},
		},
	}
	fors := CommandMap{
		"window": CommandWrap{
			minIn: 1, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				cid, err := parseClient(args[0], ctx, tracker)
				if err != nil {
					err = fmt.Errorf("Parse error for client id \"%v\": %w", args[0], err)
				}

				ctx.TargetClient = cid

				return nil, err
			},
		},
		"workspace": CommandWrap{
			minIn: 1, maxIn: 1,
			fn: func(args ...string) ([]string, error) {
				workspaceNum, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					err = fmt.Errorf("Parse error for workspace number \"%v\": %w", args[0], err)
				} else if workspaceNum >= uint64(tracker.WorkspaceCount()) {
					err = fmt.Errorf("Parse error for workspace number \"%v\": number is out of range", args[0])
				}

				ctx.TargetWorkspaceNum = uint(workspaceNum)

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

func parseClient(arg string, ctx *CommandContext, tr Tracker) (Client, error) {
	var client Client
	var err error = nil

	switch {
	case arg == "%target":
		client = ctx.TargetClient
	case arg == "%queried":
		fallthrough
	case strings.HasPrefix(arg, "%"):
		var exists bool
		client, exists = ctx.Variables[arg]
		if !exists {
			err = fmt.Errorf("Parse error for client id \"%v\": variable do not exists", arg)
		}
	default:
		var exists bool
		id, err := tr.ParseClientId(arg)
		if err != nil {
			return nil, fmt.Errorf("Parse error for client id \"%v\": %w", arg, err)
		}
		client, exists = tr.Client(id)
		if !exists {
			err = fmt.Errorf("Client id \"%v\" is not tracked", arg)
		}
	}

	return client, err
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
