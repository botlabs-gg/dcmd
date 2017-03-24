package dcmd

import (
	"strings"
)

type MiddleWareFunc func(next RunFunc) RunFunc
type RunFunc func(data *Data) (interface{}, error)

// Container is the standard muxer
// Containers can be nested by calling Container.Sub(...)
type Container struct {
	// Default mention handler, used when the bot is mentioned without any command specified
	DefaultMention Cmd

	// Default not found handler, called when no command is found from input
	NotFound Cmd

	// Default DM not found handler, same as NotFound but for Direct Messages, if none specified
	// will use notfound if set.
	DMNotFound Cmd

	// Set to ignore bots
	IgnoreBots bool
	// Dumps the stack in a response message when a panic happens in a command
	SendStackOnPanic bool
	// Set to send error messages that a command returned as a response message
	SendError bool
	// Set to also run this muxer in dm's
	RunInDM bool

	// The muxer names
	names []string
	// The muxer description
	Description string
	// The muxer long description
	LongDescription string

	// Commands this muxer will check
	Commands []Cmd

	// Hooks to be ran before executing the command
	// if the hook returns false, it will not execute any hooks or the command itself after it
	middlewares []MiddleWareFunc

	HelpTitleEmoji string
	HelpColor      int
	HelpOwnEmbed   bool
	Category       string

	Parent *Container
}

var (
	_ Cmd                 = (*Container)(nil)
	_ CmdWithDescriptions = (*Container)(nil)
)

func (c *Container) Names() []string                { return c.names }
func (c *Container) Descriptions() (string, string) { return c.Description, c.LongDescription }

func (c *Container) Run(data *Data) (interface{}, error) {
	if c.shouldIgnore(data) {
		return nil, nil
	}

	matchingCmd, rest := c.FindCommand(data.MsgStrippedPrefix)

	data.ContainerChain = append(data.ContainerChain, c)

	if matchingCmd == nil {
		var defaultHandler Cmd
		if data.MsgStrippedPrefix == "" && data.Source == MentionSource && c.DefaultMention != nil {
			defaultHandler = c.DefaultMention
		} else if data.Source == MentionSource || data.Source == PrefixSource {
			defaultHandler = c.NotFound
		} else if data.Source == DMSource {
			defaultHandler = c.DMNotFound
		}
		if defaultHandler != nil {
			return defaultHandler.Run(data)
		}

		// No handler to run, do nothing...
		return nil, nil
	}

	data.MsgStrippedPrefix = rest
	data.Cmd = matchingCmd

	if _, ok := matchingCmd.(*Container); ok {
		return matchingCmd.Run(data)

	}

	// Build the run chain
	var last RunFunc = matchingCmd.Run
	for i := range data.ContainerChain {
		last = data.ContainerChain[len(data.ContainerChain)-1-i].buildMiddlewareChain(last)
	}

	return last(data)
}

func (c *Container) shouldIgnore(data *Data) bool {
	if c.IgnoreBots && data.Msg.Author.Bot {
		return true
	}

	if data.Source == DMSource && !c.RunInDM {
		return true
	}

	return false
}

func (c *Container) FindCommand(searchStr string) (cmd Cmd, rest string) {
	split := strings.SplitN(searchStr, " ", 2)
	if len(split) < 1 {
		return
	}

	// Start looking for matches in all subcommands
	for _, c := range c.Commands {
		names := c.Names()
		for _, name := range names {
			if !strings.EqualFold(name, split[0]) {
				continue
			}

			// found match!
			cmd = c
			rest = strings.TrimSpace(searchStr[len(name):])

			return
		}
	}

	// No command found
	return nil, searchStr
}

func (c *Container) AbsFindCommand(searchStr string) (cmd Cmd, container *Container) {
	container = c
	if searchStr == "" {
		return
	}

	for {
		cmd, searchStr = c.FindCommand(searchStr)
		if cmd == nil {
			return
		}

		if cast, ok := cmd.(*Container); ok {
			return cast.AbsFindCommand(searchStr)
		}

		return
	}

	return
}

// Sub returns a copy of the container but with the following attributes overwritten
// and no commands registered
func (c *Container) Sub(names ...string) *Container {
	cop := new(Container)
	*cop = *c

	cop.Commands = nil
	cop.names = names
	cop.Description = ""
	cop.LongDescription = ""
	cop.Parent = c

	c.AddCommands(cop)

	return cop
}

func (c *Container) AddCommands(cmds ...Cmd) {
	c.Commands = append(c.Commands, cmds...)
}

func (c *Container) AddMidlewares(mw ...MiddleWareFunc) {
	c.middlewares = append(c.middlewares, mw...)
}

func (c *Container) buildMiddlewareChain(r RunFunc) RunFunc {
	for i := range c.middlewares {
		r = c.middlewares[len(c.middlewares)-1-i](r)
	}

	return r
}

func (c *Container) FullName(aliases bool) string {
	name := ""
	if c.Parent != nil {
		name = c.Parent.FullName(aliases)
	}

	if len(c.names) < 1 {
		return name
	}

	if name != "" {
		name += " "
	}

	for i, v := range c.names {
		if i != 0 && !aliases {
			return name
		}
		if i != 0 {
			name += "/"
		}

		name += v
	}

	return name
}
