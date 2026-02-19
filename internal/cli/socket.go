package cli

import (
	"github.com/Alnivel/zentile/internal/socket"
	"github.com/Alnivel/zentile/internal/types"
)

type CommandResult struct {
	command *socket.Message
	reply   *socket.Message
	err     error
}

func sendCommands(socketPath string, commands []types.Command) (<-chan CommandResult, error) {
	c, err := socket.Dial(socketPath)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan CommandResult)
	go func() {
		defer c.Close()
		defer close(resultChan)

		for _, command := range commands {
			commandMessage := socket.Message{
				Kind: string(command.Kind),
				Args: append([]string{command.Name}, command.Args...),
			}

			err := c.SendMessage(commandMessage)
			if err != nil {
				resultChan <- CommandResult{&commandMessage, nil, err}
				return
			}

			replyMessage, err := c.Receive()
			if err != nil {
				resultChan <- CommandResult{&commandMessage, nil, err}
				return
			}

			resultChan <- CommandResult{&commandMessage, &replyMessage, nil}
		}
	}()

	return resultChan, nil
}
