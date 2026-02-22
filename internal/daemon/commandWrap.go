package daemon

import "fmt"

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
