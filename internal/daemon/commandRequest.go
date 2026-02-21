package daemon

import "github.com/Alnivel/zentile/internal/types"

type CommandRequest struct {
	Command   types.Command
	replyChan chan<- types.CommandResult
}

func NewCommandRequest(command types.Command) (CommandRequest, <-chan types.CommandResult) {
	replyChan := make(chan types.CommandResult)
	return CommandRequest{
		Command:   command,
		replyChan: replyChan,
	}, replyChan
}

func (r CommandRequest) SendResult(result types.CommandResult) {
	r.replyChan <- result
}

var startCommandSequenceCommand = types.Command{
	Kind: types.Action,
	Name: "__start_new_command_sequence",
	Args: nil,
}

func requestStartNewCommandSequence(commandChan chan<- CommandRequest) {
	request, replyChan := NewCommandRequest(startCommandSequenceCommand)
	commandChan <- request
	_ = <-replyChan
}
