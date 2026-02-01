package cli

import (
	"errors"
	"fmt"
	"iter"
	"slices"

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
			command, err = parseSetter(currentArg, getNextArg)
		case isQuery(currentArg):
			command, err = parseQuery(currentArg, getNextArg)
		case isAction(currentArg):
			command, err = parseAction(currentArg, getNextArg)
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

func isSetter(currentArg string) bool {
	return currentArg == "set"
}

func parseSetter(currentArg string, getNextArg func() (string, bool)) (Command, error) {
	_ = currentArg
	setterName, exists := getNextArg()
	if exists == false {
		return Command{}, fmt.Errorf(
			"%w: what to set is not provided",
			TooFewArguments,
		)
	}

	if _, exists = daemonCommands.Setters[setterName]; exists == false {
		return Command{}, fmt.Errorf(
			"%w: set %v",
			UnknownCommand, setterName,
		)
	}

	setterArgs, success := pullNArgs(getNextArg, 1)
	if success == false {
		return Command{}, fmt.Errorf(
			"%w: set %v expects %v argument, but %v was provided",
			TooFewArguments, setterName, 1, len(setterArgs),
		)
	}

	return Command{
		Kind: "SET",
		Args: append([]string{setterName}, setterArgs...),
	}, nil
}

func isQuery(currentArg string) bool {
	return currentArg == "query"
}

func parseQuery(currentArg string, getNextArg func() (string, bool)) (Command, error) {
	_ = currentArg
	queryName, exists := getNextArg()
	if exists == false {
		return Command{}, fmt.Errorf(
			"%w: what to query is not provided",
			TooFewArguments,
		)
	}

	if _, exists = daemonCommands.Queries[queryName]; exists == false {
		return Command{}, fmt.Errorf(
			"%w: query %v",
			UnknownCommand, queryName,
		)
	}

	if err := checkCommandExists("query", queryName); err != nil {
		return Command{}, err
	}

	return Command{
		Kind: "QUERY",
		Args: []string{queryName},
	}, nil
}

func isAction(currentArg string) bool {
	_, exists := daemonCommands.Actions[currentArg]
	return exists
}

func parseAction(currentArg string, getNextArg func() (string, bool)) (Command, error) {
	_ = getNextArg

	if err := checkCommandExists("action", currentArg); err != nil {
		return Command{}, err
	}

	return Command{
		Kind: "ACTION",
		Args: []string{currentArg},
	}, nil
}

func checkCommandExists(kind string, name string) error {
	var exists bool = false

	switch kind {
	case "query":
		map_ := daemonCommands.Queries
		_, exists = map_[name]
	case "set":
		map_ := daemonCommands.Setters
		_, exists = map_[name]
	case "action":
		map_ := daemonCommands.Actions
		_, exists = map_[name]
	}

	if !exists {
		return fmt.Errorf(
			"%w: %v %v",
			UnknownCommand, kind, name,
		)
	}
	return nil
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
