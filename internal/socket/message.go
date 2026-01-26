package socket

import (
	"fmt"
	"strings"
)

type Message struct {
	Kind string
	Args []string
}


func (message *Message) String() string {
	return fmt.Sprintf("%s: %s", message.Kind, strings.Join(message.Args, ", "))
}
