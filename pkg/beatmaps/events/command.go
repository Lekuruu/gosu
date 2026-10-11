package events

type CommandType string

const (
	CommandFade        CommandType = "F"
	CommandMove        CommandType = "M"
	CommandMoveX       CommandType = "MX"
	CommandMoveY       CommandType = "MY"
	CommandScale       CommandType = "S"
	CommandVectorScale CommandType = "V"
	CommandRotate      CommandType = "R"
	CommandColour      CommandType = "C"
	CommandParameter   CommandType = "P"
	CommandLoop        CommandType = "L"
	CommandTrigger     CommandType = "T"
)

type Command struct {
	Type       CommandType
	Parameters []string
	Children   []*Command
}

func (command CommandType) IsBlock() bool {
	return command == CommandLoop || command == CommandTrigger
}

func parseCommandType(value string) (CommandType, bool) {
	command := CommandType(value)
	switch command {
	case CommandFade,
		CommandMove,
		CommandMoveX,
		CommandMoveY,
		CommandScale,
		CommandVectorScale,
		CommandRotate,
		CommandColour,
		CommandParameter,
		CommandLoop,
		CommandTrigger:
		return command, true
	default:
		return "", false
	}
}
