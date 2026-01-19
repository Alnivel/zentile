package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alnivel/zentile/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	OK            = 0
	COMMAND_ERROR = 1
	PARSE_ERROR   = 2
	SOCKET_ERROR  = 3
)

func Run(config config.Config, args []string) {
	statusCode := OK
	commandsStatusCode := OK

	defer func() {
		if statusCode == OK {
			statusCode = commandsStatusCode
		}
		os.Exit(statusCode)
	}()

	commands, err := parseArgs(args)
	if err != nil {
		log.Error(err.Error())
		statusCode = PARSE_ERROR
		return
	}

	socketPath := "/tmp/zentile.sock"
	resultChan, err := sendCommands(socketPath, commands)
	if err != nil {
		log.Error(err.Error())
		statusCode = SOCKET_ERROR
		return
	}

	for r := range resultChan {
		const logFormat = "\n\t> %s\n\t< %s\n"

		switch {
		case r.err != nil:
			statusCode = SOCKET_ERROR
			log.Errorf(logFormat, r.command, r.err)
		case r.reply.Kind == "ERR":
			commandsStatusCode = COMMAND_ERROR
			log.Errorf(logFormat, r.command, r.reply)
		default:
			log.Debugf(logFormat, r.command, r.reply)

			if len(r.reply.Args) > 0 {
				formatedResultArgs := strings.Join(r.reply.Args, ", ")
				fmt.Println(formatedResultArgs)
			}
		}

	}
}
