package daemon

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type SocketCommand struct {
	name      string
	replyChan chan<- string
}

func (command SocketCommand) ReplyOk() {
	command.replyChan <- "OK"
	close(command.replyChan)
}

func (command SocketCommand) ReplyErr(msg string) {
	command.replyChan <- "ERR: " + msg
	close(command.replyChan)
}

func ListenSocket() (<-chan SocketCommand, error) {
	socketPath := "/tmp/echo.sock"

	// Unlink the socket file before binding to prevent "address already in use" errors
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, err
		}
	}

	// Listen on the Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err == nil {
		return nil, err
	}
	defer listener.Close()

	// Handle graceful shutdown signals to ensure the socket file is removed
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal) {
		<-c
		fmt.Println("Caught signal: shutting down and removing socket.")
		listener.Close()
		os.Exit(0)
	}(sigc)

	commandChan := make(chan SocketCommand)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			break
		}
		go handleConnection(conn, commandChan)
	}
	return commandChan, nil

}

func handleConnection(conn net.Conn, commandChan chan<- SocketCommand) {
	defer conn.Close()

	readBuf := make([]byte, 512)
	commandBuf := make([]byte, 0, 512)
ReadLoop:
	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			// TODO: Add logger
			return
		}

		for split := range bytes.SplitSeq(readBuf, []byte{0}) {
			n -= len(split) + 1 // Account for the terminating null byte

			if cap(commandBuf) < len(commandBuf)+len(split) {
				err := sendReplyOutside(conn, "ERR: You talking too long")
				if err != nil {
					// TODO: Add logger
					return
				}
			}

			commandBuf = append(commandBuf, split...)

			if n < 0 { // If there was no terminating null byte
				// i.e. the command is split between reads
				continue ReadLoop
				// The current split is the last in the current readBuf
				// So breaking inner or continuing outer loop is the same
			}

			commandName := string(commandBuf)
			commandBuf = commandBuf[:0] // Reset the buffer

			reply := sendCommandInside(commandChan, commandName)

			err := sendReplyOutside(conn, reply)
			if err != nil {
				// TODO: Add logger
				return
			}

		}
	}
}

func sendCommandInside(commandChan chan<- SocketCommand, commandName string) (reply string) {
	replyChan := make(chan string)
	commandChan <- SocketCommand{name: commandName, replyChan: replyChan}
	return <-replyChan
}

func sendReplyOutside(conn net.Conn, reply string) error {
	terminatedReply := append([]byte(reply), 0)

	_, err := conn.Write(terminatedReply)
	return err
}
