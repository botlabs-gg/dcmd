// Package dcmd provides a command system for use with discord bots
package dcmd

type contextKey int

const (
	// This key holds the message, both stripped from the `KeyStrippedMessageFromCommands` and also stripped from all switches, if the command implements the
	// `CmdWithSwitches` interface
	KeyPrefix contextKey = iota
)
