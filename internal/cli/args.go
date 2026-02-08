package cli

import (
	"errors"
	"fmt"
	"iter"
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

const COMMAND_SEPARATOR = ","

func parseArgs(args []string) ([]Command, error) {
	getNextArg, stopArgIter := iter.Pull(resplitSeq(args, COMMAND_SEPARATOR, splitBeforeSeq))
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
		case currentArg == COMMAND_SEPARATOR:
			continue
		case isSetter(currentArg):
			fallthrough
		case isQuery(currentArg):
			kind := currentArg

			currentArg, exists := getNextArg()
			if !exists || currentArg == COMMAND_SEPARATOR {
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

	commandArgs, _ := pullUntilSepOrN(getNextArg, COMMAND_SEPARATOR, command.MaxIn)
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

func pullUntilSepOrN[T comparable](getNextElem func() (T, bool), sep T, count int) ([]T, bool) {
	result := make([]T, 0, count)
	for range count {
		value, exists := getNextElem()
		if !exists || value == sep {
			return result, false
		}

		result = append(result, value)
	}

	return result, true
}

func resplitSeq[T any](it []T, sep T, splitFn func(s, sep T) iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, el := range it {
			for splitEl := range splitFn(el, sep) {
				if !yield(splitEl) {
					return
				}
			}
		}
	}
}

func splitBeforeSeq(s, sep string) iter.Seq[string] {
	return func(yield func(string) bool) {
		sepLength := len(sep)
		for {
			if len(s) == 0 {
				return
			}

			sepIndex := strings.Index(s, sep)

			if sepIndex == -1 {
				if !yield(s) {
					return
				}
				break
			}

			if sepIndex != 0 {
				if !yield(s[:sepIndex]) {
					return
				}
			}

			if !yield(sep) {
				return
			}

			s = s[sepIndex+sepLength:]
		}
	}
}
