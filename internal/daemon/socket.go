package daemon

import (
	"errors"
	"io"

	"github.com/Alnivel/zentile/internal/socket"
	"github.com/Alnivel/zentile/internal/types"
	log "github.com/sirupsen/logrus"
)

type Listener struct {
	socket.Listener
}

func ListenSocket(path string) (Listener, error) {
	listener, err := socket.Listen(path)
	return Listener{listener}, err
}

func (listener Listener) HandleIncomingCommands() (<-chan types.Command, chan<- types.CommandResult) {
	commandChan := make(chan types.Command)
	commandResultChan := make(chan types.CommandResult)
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Warningf("Accept error: %v\n", err)
				return
			}
			go handleConnection(conn, commandChan, commandResultChan)
		}
	}()
	return commandChan, commandResultChan

}


func handleConnection(conn socket.Conn, commandChan chan<- types.Command, commandResultChan <-chan types.CommandResult) {
	defer conn.Close()
	log.Debug("Connection accepted")

	for {
		message, err := conn.Receive()
		switch err {
		case socket.ReadError:
			// TODO: Add logger
			return
		case socket.SplitTooLongError:
			err := conn.Send("ERR", "You talking too long")
			_ = err
			// TODO: Add logger
			return
		}


		switch message.Kind {
		case "PING":
			conn.Send("PONG")
		case "ACTION":
			fallthrough
		case "SET":
			fallthrough
		case "QUERY":
			if len(message.Args) >= 1 {
				commandChan <- types.Command{
					Kind: types.CommandType(message.Kind),
					Name: message.Args[0],
					Args: message.Args[1:],
				}

				result := <-commandResultChan
				sendErrOrVals(conn, result.Err, result.Messages...)
			} else {
				conn.Send("ERR", "Command must have at least one argument")
			}
		}


		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug("Connection was closed")
			} else {
				log.Warningf(
					"Error during handling command:\n\t%s\n\t%s\n",
					message,
					err.Error(),
				)
			}

			return
		}
	}
}

func sendErrOrVals(conn socket.Conn, err error, vals ...string) {
	if err == nil {
		conn.Send("OK", vals...)
	} else {
		conn.Send("ERR", err.Error())
	}
}
