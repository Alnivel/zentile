package socket

import (
	"iter"
	"net"
	"strconv"
)

type Conn struct {
	conn         net.Conn
	readIterNext func() ([]byte, error, bool)
	readIterStop func()
}

func Dial(path string) (Conn, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return Conn{}, err
	}

	return NewSocketConn(conn), nil
}

func NewSocketConn(conn net.Conn) Conn {
	readIterNext, readIterStop := iter.Pull2(readSplitSeq(conn, []byte{0}))

	return Conn{
		conn:         conn,
		readIterNext: readIterNext,
		readIterStop: readIterStop,
	}
}

func (s *Conn) Close() {
	s.readIterStop()
	s.conn.Close()

}

func (s *Conn) Read() (string, error) {
	splitBuf, error, done := s.readIterNext()
	_ = done
	return string(splitBuf), error
}

func (s *Conn) Write(str string) error {
	terminatedStr := append([]byte(str), 0)

	_, err := s.conn.Write(terminatedStr)
	return err
}

func (s *Conn) Send(messageKind string, messageArgs ...string) error {
	message := make([]byte, 0, 512)

	message = append(message, []byte(messageKind)...)
	message = append(message, 0)

	message = append(message, strconv.Itoa(len(messageArgs))...)
	message = append(message, 0)

	for _, messageArg := range messageArgs {
		message = append(message, []byte(messageArg)...)
		message = append(message, 0)
	}

	_, err := s.conn.Write(message)
	return err
}

func (s *Conn) SendMessage(message Message) error {
	return s.Send(message.Kind, message.Args...)
}

func (s *Conn) Receive() (Message, error) {
	kind, err := s.Read()
	if err != nil {
		return Message{}, err
	}

	argcStr, err := s.Read()
	if err != nil {
		return Message{}, err
	}

	argc, err := strconv.Atoi(argcStr)
	if err != nil {
		return Message{}, err
	}

	args := make([]string, argc)
	for i := range argc {
		args[i], err = s.Read()
		if err != nil {
			return Message{}, err
		}
	}

	return Message{
		Kind: kind,
		Args: args,
	}, err
}
