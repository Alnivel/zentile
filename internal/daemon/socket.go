package daemon

import (
	"errors"
	"io"

	"github.com/Alnivel/zentile/internal/socket"
	log "github.com/sirupsen/logrus"
)

type Listener struct {
	socket.Listener
}

func ListenSocket(path string) (Listener, error) {
	listener, err := socket.Listen(path)
	return Listener{listener}, err
}

func (listener Listener) HandleIncomingCommands(commands Commands) (<-chan struct{}, <-chan struct{}) {
	pingBefore := make(chan struct{})
	pingAfter := make(chan struct{})
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Warningf("Accept error: %v\n", err)
				return
			}
			go handleConnection(conn, commands, pingBefore, pingAfter)
		}
	}()
	return pingBefore, pingAfter

}

func handleConnection(conn socket.Conn, commands Commands, pingBefore chan<- struct{}, pingAfter chan<- struct{}) {
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

		pingBefore <- struct{}{}

		switch message.Kind {
		case "PING":
			conn.Send("PONG")
		case "ACTION":
			name := message.Args[0]
			err = commands.DoAction(name)
			sendErrOrVals(conn, err)
		case "QUERY":
			name := message.Args[0]
			result, err := commands.DoQuery(name)
			sendErrOrVals(conn, err, result)
		}

		pingAfter <- struct{}{}

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
