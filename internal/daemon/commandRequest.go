package daemon

import "github.com/Alnivel/zentile/internal/types"

type CommandRequest struct {
	Command types.Command
	replyChan chan<-types.CommandResult
}

func NewCommandRequest(command types.Command) (CommandRequest, <-chan types.CommandResult) {
	replyChan := make(chan types.CommandResult)
	return CommandRequest{
		Command: command,
		replyChan: replyChan,
	}, replyChan
}

func (r CommandRequest) SendResult(result types.CommandResult) {
	r.replyChan <- result
}
