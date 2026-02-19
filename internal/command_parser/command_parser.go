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

// Parse provided string into slice of commands. 
// The commnad string are split by spaces and commas. 
// A command implicitly takes up to the maximal valid number of arguments for the command
// while comma explicitly denotes the end of arguments for the command.
func (parser CommandParser) ParseString(s string) ([]types.Command, error) {
	return parser.ParseSeq(strings.SplitSeq(s, " "))
}

// Parse provided slice of strings into slice of commands.
// The strings are pressumed to be already being split by space and resplit only by comma.
// A command implicitly takes up to the maximal valid number of arguments for the command
// while comma explicitly denotes the end of arguments for the command.
func (parser CommandParser) ParseSlice(s []string) ([]types.Command, error) {
	return parser.ParseSeq(slices.Values(s))
}

// Parse provided iterator over strings into slice of commands.
// The strings are pressumed to be already being split by space and resplit only by comma.
// A command implicitly takes up to the maximal valid number of arguments for the command
// while comma explicitly denotes the end of arguments for the command.
func (parser CommandParser) ParseSeq(it iter.Seq[string]) ([]types.Command, error) {
	getNextToken, stopTokenIter := iter.Pull(
		resplitSeq( it, splitAtSeq, COMMAND_SEPARATOR,
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

// Parse individual command
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

// Call getNextElem function n times or until
// the function returns false or sep.
// Returns slice of the collected values and
// status if all requested n values was collected
func pullUntilSepOrN[T comparable](getNextElem func() (T, bool), sep T, n int) ([]T, bool) {
	result := make([]T, 0, n)
	for range n {
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
