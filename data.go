package dcmd

import (
	"context"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/dstate"
)

type Data struct {
	Cmd      *RegisteredCommand
	Args     []*ParsedArg
	Switches map[string]*ParsedArg

	Msg     *discordgo.Message
	CS      *dstate.ChannelState
	GS      *dstate.GuildState
	Session *discordgo.Session
	Source  TriggerSource

	PrefixUsed string

	// The message with the prefix removed (either mention or command prefix)
	MsgStrippedPrefix string

	// The chain of containers we went through, first element is always root
	ContainerChain []*Container

	// The system that triggered this command
	System *System

	context context.Context
}

// Context returns an always non-nil context
func (d *Data) Context() context.Context {
	if d.context == nil {
		return context.Background()
	}

	return d.context
}

func (d *Data) Switch(name string) *ParsedArg {
	return d.Switches[name]
}

// WithContext creates a copy of d with the context set to ctx
func (d *Data) WithContext(ctx context.Context) *Data {
	cop := new(Data)
	*cop = *d
	cop.context = ctx
	return cop
}

// Where this command comes from
type TriggerSource int

const (
	DMSource TriggerSource = iota
	MentionSource
	PrefixSource
)
