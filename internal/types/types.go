package types

type CommandType string

const (
	Action CommandType = "ACTION"
	Set    CommandType = "SET"
	Query  CommandType = "QUERY"
	For    CommandType = "FOR"
)

type Command struct {
	Kind CommandType
	Name string
	Args []string
}

type CommandResult struct {
	Messages []string
	Err      error
}

