package socket

import (
	"bytes"
	"errors"
	"iter"
	"net"
)

var (
	ReadError         = errors.New("Failed to read from connection")
	SplitTooLongError = errors.New("Message is longer than the read buffer")
	InvalidSeparatorError = errors.New("Separator cannot be zero length")
)

func readSplitSeq(conn net.Conn, sep []byte) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		readBuf := make([]byte, 512)
		splitBuf := make([]byte, 0, 512)

		sepLen := len(sep)
		if sepLen == 0 {
			yield(nil, InvalidSeparatorError)
			return
		}
	ReadLoop:
		for {
			n, err := conn.Read(readBuf)
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			for split := range bytes.SplitSeq(readBuf[:n], sep) {
				n -= len(split) + sepLen // Account for the separator

				if cap(splitBuf) < len(splitBuf)+len(split) {
					if !yield(nil, SplitTooLongError) {
						return
					}
				}

				splitBuf = append(splitBuf, split...)

				if n < 0 { // If there was no separator
					// i.e. the command is split between reads
					continue ReadLoop
					// The current split is the last in the current readBuf
					// So breaking inner or continuing outer loop is the same
				}

				if !yield(splitBuf, nil) {
					return
				}
				splitBuf = splitBuf[:0] // Reset the buffer
			}
		}
	}
}

