package daemon

import (
	"errors"
	"io"
	"sync"

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

func (listener Listener) HandleIncomingCommands(commandChan chan<- CommandRequest, chanMutex *sync.Mutex) {
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Warningf("Accept error: %v\n", err)
				return
			}
			go handleConnection(conn, commandChan, chanMutex)
		}
	}()
}

func handleConnection(conn socket.Conn, commandChan chan<- CommandRequest, chanMutex *sync.Mutex) {
	defer conn.Close()
	log.Debug("Connection accepted")

	chanMutex.Lock()
	defer chanMutex.Unlock()
	requestStartNewCommandSequence(commandChan)

	for {
		var errOnReceive, errOnSend error
		message, errOnReceive := conn.Receive()

		switch errOnReceive {
		case socket.ReadError:
			log.Error(errOnReceive.Error())
			return
		case socket.SplitTooLongError:
			log.Error(errOnReceive.Error())

			errOnSend := conn.Send("ERR", "You talking too long")
			log.Errorf("Failed to send error to the client: %v\n", errOnSend)
			return
		}

		switch message.Kind {
		case "PING":
			errOnSend = conn.Send("PONG")
		case "ACTION":
			fallthrough
		case "SET":
			fallthrough
		case "QUERY":
			fallthrough
		case "FOR":
			if len(message.Args) >= 1 {
				command := types.Command{
					Kind: types.CommandType(message.Kind),
					Name: message.Args[0],
					Args: message.Args[1:],
				}
				commandRequest, replyChan := NewCommandRequest(command)
				commandChan <- commandRequest

				result := <-replyChan
				errOnSend = sendCommandResult(conn, command, result)
			} else {
				errOnSend = conn.Send("ERR", "Command must have at least one argument")
			}
		}

		if errOnReceive != nil {
			logProtocolErr(errOnReceive, message)
			return
		}
		if errOnSend != nil {
			logProtocolErr(errOnSend, message)
			return
		}
	}
}

func logProtocolErr(err error, message socket.Message) {
	if errors.Is(err, io.EOF) {
		log.Debug("Connection was closed")
	} else {
		log.Warningf(
			"Error during handling command:\n\t%s\n\t%s\n",
			message,
			err.Error(),
		)
	}
}

func sendCommandResult(conn socket.Conn, command types.Command, result types.CommandResult) error {
	if result.Err == nil {
		return conn.Send("OK", result.Messages...)
	} else {
		log.Warningf(
			"Error during handling command:\n\t%v\n\t%v\n",
			command,
			result.Err,
		)
		return conn.Send("ERR", result.Err.Error())
	}
}
