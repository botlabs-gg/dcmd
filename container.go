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
	DefaultMention RunFunc

	// Default not found handler, called when no command is found from input
	NotFound RunFunc

	// Default DM not found handler, same as NotFound but for Direct Messages, if none specified
	// will use notfound if set.
	DMNotFound RunFunc

	// Set to ignore bots
	IgnoreBots bool
	// Dumps the stack in a response message when a panic happens in a command
	SendStackOnPanic bool
	// Set to send error messages that a command returned as a response message
	SendError bool
	// Set to also run this muxer in dm's
	RunInDM bool

	// The muxer names
	Names []string
	// The muxer description
	Description string
	// The muxer long description
	LongDescription string

	// Commands this muxer will check
	Commands []*RegisteredCommand

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

func (c *Container) Descriptions(data *Data) (string, string) { return c.Description, c.LongDescription }

func (c *Container) Run(data *Data) (interface{}, error) {
	if c.shouldIgnore(data) {
		return nil, nil
	}

	matchingCmd, rest := c.FindCommand(data.MsgStrippedPrefix)

	data.ContainerChain = append(data.ContainerChain, c)

	if matchingCmd == nil {
		var defaultHandler RunFunc
		if data.MsgStrippedPrefix == "" && data.Source == MentionSource && c.DefaultMention != nil {
			defaultHandler = c.DefaultMention
		} else if data.Source == MentionSource || data.Source == PrefixSource {
			defaultHandler = c.NotFound
		} else if data.Source == DMSource {
			defaultHandler = c.DMNotFound
		}
		if defaultHandler != nil {
			return defaultHandler(data)
		}

		// No handler to run, do nothing...
		return nil, nil
	} else {
		if matchingCmd.Trigger.DisableInDM && data.Source == DMSource {
			// Disabled in dms
			return nil, nil
		} else if matchingCmd.Trigger.DisableOutsideDM && data.Source != DMSource {
			// Disabled outside dms
			return nil, nil
		}
	}

	data.MsgStrippedPrefix = rest
	data.Cmd = matchingCmd

	if _, ok := matchingCmd.Command.(*Container); ok {
		return matchingCmd.Command.Run(data)

	}

	// Build the run chain
	var last RunFunc = matchingCmd.Command.Run

	// User either prebuilt middleware chain, or build it live
	if matchingCmd.builtFullMiddlewareChain != nil {
		last = matchingCmd.builtFullMiddlewareChain
	} else {

		for i := range matchingCmd.Trigger.Middlewares {
			last = matchingCmd.Trigger.Middlewares[len(matchingCmd.Trigger.Middlewares)-1-i](last)
		}

		for i := range data.ContainerChain {
			last = data.ContainerChain[len(data.ContainerChain)-1-i].buildMiddlewareChain(last, matchingCmd)
		}
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

func (c *Container) FindCommand(searchStr string) (cmd *RegisteredCommand, rest string) {
	split := strings.SplitN(searchStr, " ", 2)
	if len(split) < 1 {
		return
	}

	// Start looking for matches in all subcommands
	for _, c := range c.Commands {
		names := c.Trigger.Names
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

func (c *Container) AbsFindCommand(searchStr string) (cmd *RegisteredCommand, container *Container) {
	cmd, container, _ = c.AbsFindCommandWithRest(searchStr)
	return
}

func (c *Container) AbsFindCommandWithRest(searchStr string) (cmd *RegisteredCommand, container *Container, rest string) {
	container = c
	if searchStr == "" {
		return
	}

	for {
		cmd, searchStr = c.FindCommand(searchStr)
		rest = searchStr
		if cmd == nil {
			return
		}

		if cast, ok := cmd.Command.(*Container); ok {
			return cast.AbsFindCommandWithRest(searchStr)
		}

		return
	}

	return
}

// Sub returns a copy of the container but with the following attributes overwritten
// and no commands registered
func (c *Container) Sub(mainName string, aliases ...string) *Container {
	cop := new(Container)
	*cop = *c

	cop.Commands = nil
	cop.Names = append([]string{mainName}, aliases...)
	cop.Description = ""
	cop.LongDescription = ""
	cop.middlewares = nil
	cop.Parent = c

	c.AddCommand(cop, NewTrigger(mainName, aliases...))

	return cop
}

func (c *Container) AddCommand(cmd Cmd, trigger *Trigger) *RegisteredCommand {
	wrapped := &RegisteredCommand{
		Command: cmd,
		Trigger: trigger,
	}

	c.Commands = append(c.Commands, wrapped)
	return wrapped
}

func (c *Container) AddMidlewares(mw ...MiddleWareFunc) {
	c.middlewares = append(c.middlewares, mw...)
}

func (c *Container) buildMiddlewareChain(r RunFunc, cmd *RegisteredCommand) RunFunc {
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

	if len(c.Names) < 1 {
		return name
	}

	if name != "" {
		name += " "
	}

	for i, v := range c.Names {
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

// BuildMiddlewareChains builds all the middleware chains and chaches them.
// It is reccomended to call this after adding all commands and middleware to avoid building the chains everytime
// a command is invoked
func (c *Container) BuildMiddlewareChains(containerChain []*Container) {
	containerChain = append(containerChain, c)
	for _, cmd := range c.Commands {
		if cast, ok := cmd.Command.(*Container); ok {
			cast.BuildMiddlewareChains(containerChain)
			continue
		}

		last := cmd.Command.Run

		for i := range cmd.Trigger.Middlewares {
			last = cmd.Trigger.Middlewares[len(cmd.Trigger.Middlewares)-1-i](last)
		}

		for i := range containerChain {
			last = containerChain[len(containerChain)-1-i].buildMiddlewareChain(last, cmd)
		}
		cmd.builtFullMiddlewareChain = last
	}
}
