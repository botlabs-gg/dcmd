package dcmd

type contextKey int

const (
	// This key holds the message, both stripped from the `KeyStrippedMessageFromCommands` and also stripped from all switches, if the command implements the
	// `CmdWithSwitches` interface
	KeyStrippedMessageFromSwitches contextKey = iota
)
