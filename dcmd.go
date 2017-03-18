package dcmd

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type contextKey int

const (
	// This key holds the message, both stripped from the `KeyStrippedMessageFromCommands` and also stripped from all switches, if the command implements the
	// `CmdWithSwitches` interface
	KeyStrippedMessageFromSwitches contextKey = iota
)

// GenerateFullHelp generates full help for a container
func GenerateHelp(titlePrefix string, container *Container) []*discordgo.MessageEmbed {

	embeds := make([]*discordgo.MessageEmbed, 0, 1)

	str := ""

	currentContainerName := ""
	if len(container.names) > 0 {
		currentContainerName = container.names[0]
	}

	for _, v := range container.Commands {

		// The container has it's own embed
		if c, ok := v.(*Container); ok {
			pref := currentContainerName
			if titlePrefix != "" {
				pref = titlePrefix + " " + pref
			}

			embeds = append(embeds, GenerateHelp(pref, c)...)
		}

		// Add the current container stack to the name
		namesStr := ""
		if titlePrefix != "" {
			namesStr += titlePrefix
		}
		if currentContainerName != "" {
			if namesStr != "" {
				namesStr = namesStr + " " + currentContainerName + " "
			} else {
				namesStr = currentContainerName + " "
			}
		}

		// Add the command names
		for k, name := range v.Names() {
			if k != 0 {
				namesStr += "/"
			}
			namesStr += name
		}

		// Add the short description, if available
		desc := ""
		if cast, ok := v.(CmdWithDescriptions); ok {
			short, long := cast.Descriptions()
			if short != "" {
				desc = ": " + short
			} else if long != "" {
				desc = ": " + long
			}
		}

		str += fmt.Sprintf("`%s`%s\n", namesStr, desc)
	}

	cName := ""
	if len(container.names) > 0 {
		cName = container.names[0]
	}

	return append([]*discordgo.MessageEmbed{&discordgo.MessageEmbed{
		Title:       cName + " Help",
		Description: str,
	}}, embeds...)
}

// GenerateSingleHelp generates help for a single command
func GenerateSingleCmdHelp(containerChain string, cmd Cmd) *discordgo.MessageEmbed {

}

type StdHelpCommand struct {
	CmdNames          []string
	SendFullInDM      bool
	SendTargettedInDM bool
}

var (
	_ Cmd                 = (*StdHelpCommand)(nil)
	_ CmdWithDescriptions = (*StdHelpCommand)(nil)
)

func NewStdHelpCommand(names ...string) *StdHelpCommand {
	return &StdHelpCommand{
		CmdNames: names,
	}
}

func (h *StdHelpCommand) Names() []string {
	if len(h.CmdNames) < 1 {
		return []string{"help"}
	}

	return h.CmdNames
}

func (h *StdHelpCommand) Descriptions() (string, string) {
	return "Shows help for all or a specific command/container", ""
}

func (h *StdHelpCommand) Run(d *Data) (interface{}, error) {
	root := d.ContainerChain[0]
	help := GenerateHelp("", root)
	return help, nil
}
