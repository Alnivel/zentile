package cli

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/Alnivel/zentile/internal/daemon"
)

var daemonCommands daemon.Commands = daemon.InitCommands(nil)

type Command struct {
	Kind string
	Args []string
}

var (
	TooFewArguments = errors.New("Too few argument provided")
	UnknownCommand  = errors.New("Unknown command")
)

func parseArgs(args []string) ([]Command, error) {
	getNextArg, stopArgIter := iter.Pull(slices.Values(args))
	defer stopArgIter()

	commands := make([]Command, 0)
	for {
		currentArg, argExists := getNextArg()
		if argExists == false {
			break
		}

		var command Command
		var err error

		switch {
		case isSetter(currentArg):
			fallthrough
		case isQuery(currentArg):
			kind := currentArg

			currentArg, exists := getNextArg()
			if exists == false {
				return nil, fmt.Errorf(
					"%w: %v name is not provided",
					TooFewArguments, kind,
				)
			}
			name := currentArg

			command, err = parseCommand(kind, name, getNextArg)
		case isAction(currentArg):
			kind := "action"
			name := currentArg

			command, err = parseCommand(kind, name, getNextArg)
		default:
			err = UnknownCommand
		}

		if err != nil {
			return nil, err
		}

		commands = append(commands, command)
	}

	return commands, nil
}

func parseCommandType(arg string) daemon.CommandType {
	return daemon.CommandType(strings.ToUpper(arg))
}
func parseCommand(commandKind string, commandName string, getNextArg func() (string, bool)) (Command, error) {
	commandType := parseCommandType(commandKind)
	command, exists := daemonCommands.Map(commandType)[commandName]
	if exists == false {
		return Command{}, fmt.Errorf(
			"%w: %v %v",
			UnknownCommand, commandKind, commandName,
		)
	}

	commandArgs, _ := pullNArgs(getNextArg, command.MaxIn)
	if len(commandArgs) < command.MinIn {
		if command.MinIn == command.MaxIn {
			return Command{}, fmt.Errorf(
				"%w: %v %v expects %v argument, but %v was provided",
				TooFewArguments, commandKind, commandName, command.MinIn, len(commandArgs),
			)
		} else {
			return Command{}, fmt.Errorf(
				"%w: %v %v expects between %v and %v arguments, but %v was provided",
				TooFewArguments, commandKind, commandName, command.MinIn, command.MaxIn, len(commandArgs),
			)
		}
	}

	return Command{
		Kind: string(commandType),
		Args: append([]string{commandName}, commandArgs...),
	}, nil
}

func isSetter(currentArg string) bool {
	return currentArg == "set"
}

func isQuery(currentArg string) bool {
	return currentArg == "query"
}

func isAction(currentArg string) bool {
	_, exists := daemonCommands.Actions[currentArg]
	return exists
}

func pullNArgs(getNextArg func() (string, bool), count int) ([]string, bool) {
	result := make([]string, 0, count)
	for range count {
		value, exists := getNextArg()
		if !exists {
			return result, false
		}

		result = append(result, value)
	}

	return result, true
}
