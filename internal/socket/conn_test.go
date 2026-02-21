package socket_test

import (
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/Alnivel/zentile/internal/socket"
)

func TestConn_ReadWritePair(t *testing.T) {
	tests := []struct {
		name string

		input []string
		want  []string
	}{
		{"SingleInput", []string{"One"}, []string{"One"}},
		{"MultipleInputs", []string{"One", "Two", "Three"}, []string{"One", "Two", "Three"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeConnA, pipeConnB := net.Pipe()

			socketConnA := socket.NewSocketConn(pipeConnA)
			socketConnB := socket.NewSocketConn(pipeConnB)

			// Write in background
			go func() {
				defer socketConnA.Close()
				for _, str := range tt.input {
					writeErr := socketConnA.Write(str)
					if writeErr != nil {
						t.Errorf("Got write error %v, expecting none", writeErr)
					}
				}
			}()

			// Read in foreground
			func() {
				defer socketConnB.Close()
				got := make([]string, 0)

			ReadLoop:
				for {
					str, readErr := socketConnB.Read()

					switch readErr {
					case nil:
						got = append(got, str)
					case io.EOF:
						break ReadLoop
					default:
						t.Errorf("Got read error %v, expecting none", readErr)
						break ReadLoop
					}
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Read %v, want %v", got, tt.want)
				}
			}()

		})
	}
}

func TestConn_SendReceivePair(t *testing.T) {
	tests := []struct {
		name string

		input []socket.Message
		want  []socket.Message
	}{
		{
			name: "NormalMessages",
			input: []socket.Message{
				{"ACTION", []string{"One", "Two", "Three"}},
				{"QUERY", []string{"One", "Two", "Three"}},
			},
			want: []socket.Message{
				{"ACTION", []string{"One", "Two", "Three"}},
				{"QUERY", []string{"One", "Two", "Three"}},
			},
		},
		{ 
			// TODO: Document that Receive always returns Message with Args 
			// being a zero-length slice if Args of sent message is nil
			name: "EmptyMessage",
			input: []socket.Message{
				{},
				{"", nil},
				{"", []string{}},
			},
			want: []socket.Message{
				{"", []string{}},
				{"", []string{}},
				{"", []string{}},
			},
		},
		{
			name: "SecondMessageSmallerThanFirst",
			input: []socket.Message{
				{"123456789", []string{}},
				{"1234567", []string{}},
				{"NextMessage", []string{}},
			},
			want: []socket.Message{
				{"123456789", []string{}},
				{"1234567", []string{}},
				{"NextMessage", []string{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeConnA, pipeConnB := net.Pipe()

			socketConnA := socket.NewSocketConn(pipeConnA)
			socketConnB := socket.NewSocketConn(pipeConnB)

			// Write in background
			go func() {
				defer socketConnA.Close()
				for _, msg := range tt.input {
					writeErr := socketConnA.SendMessage(msg)
					if writeErr != nil {
						t.Errorf("Got write error %v, expecting none", writeErr)
					}
				}
			}()

			// Read in foreground
			func() {
				defer socketConnB.Close()
				got := make([]socket.Message, 0)

			ReadLoop:
				for {
					msg, readErr := socketConnB.Receive()

					switch readErr {
					case nil:
						got = append(got, msg)
					case io.EOF:
						break ReadLoop
					default:
						t.Errorf("Got read error %v, expecting none", readErr)
						break ReadLoop
					}
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Got %v, want %v", got, tt.want)
				}
			}()

		})
	}
}
