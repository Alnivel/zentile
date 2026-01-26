package cli

import (
	"errors"
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
	NoQueryArgumentProvided = errors.New("No query argument provided")
	UnknownQuery            = errors.New("Unknown query")
	UnknownCommand          = errors.New("Unknown command")
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

func isQuery(currentArg string) bool {
	return currentArg == "query"
}

func parseQuery(currentArg string, getNextArg func() (string, bool)) (Command, error) {
	_ = currentArg
	whatToQuery, exists := getNextArg()
	if exists == false {
		return Command{}, NoQueryArgumentProvided
	}

	if _, exists = daemonCommands.Queries[whatToQuery]; exists == false {
		return Command{}, UnknownQuery
	}

	return Command{
		Kind: "QUERY",
		Args: []string{whatToQuery},
	}, nil
}

func isAction(currentArg string) bool {
	_, exists := daemonCommands.Actions[currentArg]
	return exists
}

func parseAction(currentArg string, getNextArg func() (string, bool)) (Command, error) {
	_ = getNextArg
	return Command{
		Kind: "ACTION",
		Args: []string{currentArg},
	}, nil
}
