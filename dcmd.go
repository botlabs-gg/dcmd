package dcmd

import (
	"reflect"
	"strings"
)

type contextKey int

const (
	// This key holds the message, both stripped from the `KeyStrippedMessageFromCommands` and also stripped from all switches, if the command implements the
	// `CmdWithSwitches` interface
	KeyPrefix contextKey = iota
)

// CmdName retusn either the name returned from the Names function
func CmdName(cmd Cmd, aliases bool) string {
	if cmd == nil {
		return "Unknown"
	}

	if names := cmd.Names(); len(names) > 0 {
		if aliases {
			return strings.Join(names, "/")
		}

		return names[0]
	}

	t := reflect.TypeOf(cmd)
	return t.Name()
}
