package socket

import (
	"net"
)

type Listener struct {
	listener net.Listener
}

func Listen(path string) (Listener, error) {
	listener, err := net.Listen("unix", path)
	if err != nil {
		return Listener{}, err
	}

	return NewSocketListener(listener), nil
}

func NewSocketListener(listener net.Listener) Listener {
	return Listener{
		listener: listener,
	}
}

func (sl *Listener) Close() {
	sl.listener.Close()
}

func (sl *Listener) Accept() (Conn, error) {
	conn, err := sl.listener.Accept()
	if err != nil {
		return Conn{}, err
	}
	return NewSocketConn(conn), nil
}
