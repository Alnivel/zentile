package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alnivel/zentile/internal/command_parser"
	"github.com/Alnivel/zentile/internal/config"
	"github.com/Alnivel/zentile/internal/daemon"
	"github.com/Alnivel/zentile/internal/types"
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

	commands := daemon.InitCommands(nil)
	getCommandByNameAdapter := func(kind types.CommandType, name string) (commandparser.CommandWrap, bool) { 
		return commands.GetByName(kind, name) 
	}

	parser := commandparser.CommandParser{
		GetCommandByName: getCommandByNameAdapter,
	}
	parsedCommands, err := parser.Parse(args)
	if err != nil {
		log.Error(err.Error())
		statusCode = PARSE_ERROR
		return
	}

	socketCommands := make([]types.Command, len(parsedCommands))
	for i, v := range parsedCommands {
		socketCommands[i] = types.Command(v)
	}

	socketPath := "/tmp/zentile.sock"
	resultChan, err := sendCommands(socketPath, socketCommands)
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
