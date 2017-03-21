package dcmd

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// HelpFormatter is a interface for help formatters, for an example see StdHelpFormatter
type HelpFormatter interface {
	// Called when there is help generated for 2 or more commands
	ShortCmdHelp(cmd Cmd, container *Container) string

	// Called when help is only generated for 1 command
	// You are supposed to dump all command detilas such as arguments
	// the long description if it has one, switches and whatever else you have in mind.
	FullCmdHelp(cmd Cmd, container *Container) string
}

// SortedCommandEntry represents an entry in the SortdCommandSet
type SortedCommandEntry struct {
	Cmd       Cmd
	Container *Container
}

// SortedCommandSet groups a set of commands by either container or category
type SortedCommandSet struct {
	Commands []*SortedCommandEntry

	// Set if this is a set of commands grouped by categories
	Category *Category

	// Set if this is a container helpContainer
	Container *Container
}

func (s *SortedCommandSet) Name() string {
	if s.Category != nil {
		return s.Category.Name
	}

	return s.Container.FullName(false)
}

func (s *SortedCommandSet) Color() int {
	if s.Category != nil {
		return s.Category.EmbedColor
	}

	return s.Container.HelpColor
}

func (s *SortedCommandSet) Emoji() string {
	if s.Category != nil {
		return s.Category.HelpEmoji
	}

	return s.Container.HelpTitleEmoji
}

// SortCommands groups commands into sorted command sets
func SortCommands(closestGroupContainer *Container, cmdContainer *Container) []*SortedCommandSet {
	containers := make([]*SortedCommandSet, 0)

	for _, cmd := range cmdContainer.Commands {
		var keyCont *Container
		var keyCat *Category

		// Merge this containers generated command sets into the current one
		if c, ok := cmd.(*Container); ok {
			topGroup := closestGroupContainer
			if c.HelpOwnEmbed {
				topGroup = c
			}
			merging := SortCommands(topGroup, c)
			for _, mergingSet := range merging {
				if set := FindSortedCommands(containers, mergingSet.Category, mergingSet.Container); set != nil {
					set.Commands = append(set.Commands, mergingSet.Commands...)
				} else {
					containers = append(containers, mergingSet)
				}
			}

			continue
		}

		// Check if this command belongs to a specific category
		if catCmd, ok := cmd.(CmdWithCategory); ok {
			keyCat = catCmd.Category()
		}
		if keyCat == nil {
			keyCont = closestGroupContainer
		}

		if set := FindSortedCommands(containers, keyCat, keyCont); set != nil {
			set.Commands = append(set.Commands, &SortedCommandEntry{Cmd: cmd, Container: cmdContainer})
			continue
		}

		containers = append(containers, &SortedCommandSet{
			Commands:  []*SortedCommandEntry{&SortedCommandEntry{Cmd: cmd, Container: cmdContainer}},
			Category:  keyCat,
			Container: keyCont,
		})
	}

	return containers
}

// FindSortedCommands finds a command set by category or container
func FindSortedCommands(sets []*SortedCommandSet, cat *Category, container *Container) *SortedCommandSet {
	for _, set := range sets {
		if cat != nil && cat != set.Category {
			continue
		}

		if container != nil && container != set.Container {
			continue
		}

		return set
	}

	return nil
}

// GenerateFullHelp generates full help for a container
func GenerateHelp(d *Data, container *Container, formatter HelpFormatter) (embeds []*discordgo.MessageEmbed) {

	invoked := ""
	if d != nil && d.PrefixUsed != "" {
		invoked = d.PrefixUsed + " "
	}

	sets := SortCommands(container, container)

	for _, set := range sets {
		cName := set.Emoji() + set.Name()
		if cName != "" {
			cName += " "
		}

		embed := &discordgo.MessageEmbed{
			Title: cName + "Help",
			Color: set.Color(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Do `" + invoked + "help cmd/container` for more detailed information on a command/group of commands",
			},
		}

		for _, entry := range set.Commands {
			embed.Description += formatter.ShortCmdHelp(entry.Cmd, entry.Container)
		}

		embeds = append(embeds, embed)
	}

	return
}

type StdHelpFormatter struct {
}

var _ HelpFormatter = (*StdHelpFormatter)(nil)

func (s *StdHelpFormatter) FullCmdHelp(cmd Cmd, container *Container) string {
	return ""
}

func (s *StdHelpFormatter) ShortCmdHelp(cmd Cmd, container *Container) string {

	// Add the current container stack to the name
	nameStr := container.FullName(true)
	if nameStr != "" {
		nameStr += " "
	}

	// Add the command names
	for k, name := range cmd.Names() {
		if k != 0 {
			nameStr += "/"
		}
		nameStr += name
	}

	// Add the short description, if available
	desc := ""
	if cast, ok := cmd.(CmdWithDescriptions); ok {
		short, long := cast.Descriptions()
		if short != "" {
			desc = ": " + short
		} else if long != "" {
			desc = ": " + long
		}
	}

	return fmt.Sprintf("`%s`%s\n", nameStr, desc)
}

// GenerateSingleHelp generates help for a single command
func GenerateSingleCmdHelp(containerChain string, cmd Cmd) *discordgo.MessageEmbed {
	return nil
}

type StdHelpCommand struct {
	CmdNames          []string
	SendFullInDM      bool
	SendTargettedInDM bool

	Formatter HelpFormatter
}

var (
	_ Cmd                 = (*StdHelpCommand)(nil)
	_ CmdWithDescriptions = (*StdHelpCommand)(nil)
)

func NewStdHelpCommand(names ...string) *StdHelpCommand {
	return &StdHelpCommand{
		CmdNames:  names,
		Formatter: &StdHelpFormatter{},
	}
}

func (h *StdHelpCommand) Names() []string {
	if len(h.CmdNames) < 1 {
		return []string{"help"}
	}

	return h.CmdNames
}

func (h *StdHelpCommand) Descriptions() (string, string) {
	return "Shows short help for all commands, or a longer help for a specific command", "Shows help for all or a specific command" +
		"\nExamples: \n`help` - Shows a short summary about all commands\n`help info` - Shows a longer help message for info, can contain examples of how to use it.\nYou are currently reading the longer help message about the `help` command"
}

func (h *StdHelpCommand) Run(d *Data) (interface{}, error) {
	root := d.ContainerChain[0]
	help := GenerateHelp(d, root, h.Formatter)
	return help, nil
}
