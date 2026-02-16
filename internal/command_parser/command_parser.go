package commandparser

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/Alnivel/zentile/internal/types"
)

var (
	TooFewArguments = errors.New("Too few argument provided")
	UnknownCommand  = errors.New("Unknown command")
)

const COMMAND_SEPARATOR = ","

type CommandWrap interface {
	MinIn() int
	MaxIn() int
	ValidateArgCount(count int) error
}


type CommandParser struct {
	GetCommandByName func(kind types.CommandType, name string) (CommandWrap, bool)
}

func (parser CommandParser) Parse(s []string) ([]types.Command, error) {
	getNextToken, stopTokenIter := iter.Pull(
		resplitSeq(
			slices.Values(s),
			splitAtSeq, COMMAND_SEPARATOR,
		))
	defer stopTokenIter()

	commands := make([]types.Command, 0)
	for {
		currentToken, tokenExists := getNextToken()
		if tokenExists == false {
			break
		}

		var command types.Command
		var err error

		switch currentToken {
		case COMMAND_SEPARATOR:
			continue
		case "set":
			fallthrough
		case "query":
			kind := currentToken

			currentArg, exists := getNextToken()
			if !exists || currentArg == COMMAND_SEPARATOR {
				return nil, fmt.Errorf(
					"%w: %v name is not provided",
					TooFewArguments, kind,
				)
			}
			name := currentArg

			command, err = parser.parseCommand(kind, name, getNextToken)
		default:
			kind := "action"
			name := currentToken

			command, err = parser.parseCommand(kind, name, getNextToken)
		}

		if err != nil {
			return nil, err
		}

		commands = append(commands, command)
	}

	return commands, nil
}

func parseCommandType(arg string) types.CommandType {
	return types.CommandType(strings.ToUpper(arg))
}

func (parser CommandParser) parseCommand(commandKind string, commandName string, getNextArg func() (string, bool)) (types.Command, error) {
	commandType := parseCommandType(commandKind)
	command, exists := parser.GetCommandByName(commandType, commandName)
	if exists == false {
		return types.Command{}, fmt.Errorf(
			"%w: %v %v",
			UnknownCommand, commandKind, commandName,
		)
	}

	commandArgs, _ := pullUntilSepOrN(getNextArg, COMMAND_SEPARATOR, command.MaxIn())
	if err := command.ValidateArgCount(len(commandArgs)); err != nil {
		return types.Command{}, err
	}

	return types.Command{
		Kind: commandType,
		Name: commandName,
		Args: commandArgs,
	}, nil
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

// Apply splitFn to each element of the sequence and flatten the resulting sequence
func resplitSeq[T any](it iter.Seq[T], splitFn func(s, sep T) iter.Seq[T], sep T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for el := range it {
			for splitEl := range splitFn(el, sep) {
				if !yield(splitEl) {
					return
				}
			}
		}
	}
}

// Split string before and after separator.
// Separator becomes an element in the result sequence as well.
// Example: "a|b|c|" with separator "|" => "a", "|", "b", "|", "c", "|"
func splitAtSeq(s, sep string) iter.Seq[string] {
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
