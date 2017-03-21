package main

/*
This example provides 2 basic commands with static responses.
*/

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dcmd"
	"log"
	"os"
)

func main() {
	modCat := &dcmd.Category{
		Name:        "Moderation",
		Description: "Moderation commands",
		HelpEmoji:   "👮",
		EmbedColor:  0xdb0606,
	}

	system := dcmd.NewStandardSystem("[")
	system.Root.AddCommands(dcmd.NewStdHelpCommand("help"))
	system.Root.AddCommands(&StaticCmd{
		CmdNames: []string{"Status", "st"},
		Desc:     "Shows bot status",
		LongDesc: "Shows bot status such as uptime, and how many resources the bot uses",
	}, &StaticCmd{
		CmdNames: []string{"Info", "i"},
		Desc:     "Shows general bot information",
	}, &StaticCmd{
		CmdNames: []string{"8ball", "ball"},
		Desc:     "Ask the bot a yes/no question",
	}, &StaticCmd{
		CmdNames: []string{"Poke"},
		Desc:     "Pokes a user on your server",
	}, &StaticCmd{
		CmdNames: []string{"Warn"},
		Desc:     "Warns a user",
		Cat:      modCat,
	}, &StaticCmd{
		CmdNames: []string{"Kick"},
		Desc:     "Kicks a user",
		Cat:      modCat,
	}, &StaticCmd{
		CmdNames: []string{"Ban"},
		Desc:     "Bans a user",
		Cat:      modCat,
	}, &StaticCmd{
		CmdNames: []string{"Mute"},
		Desc:     "Mutes a user",
		Cat:      modCat,
	})

	musicContainer := system.Root.Sub("music", "m")
	musicContainer.HelpOwnEmbed = true
	musicContainer.HelpColor = 0xd60eab
	musicContainer.HelpTitleEmoji = "🎶"
	musicContainer.AddCommands(&StaticCmd{
		CmdNames: []string{"Join", "j"},
		Desc:     "Joins your current voice channel",
		LongDesc: "Makes the bot join your current voice channel, can also be used to move it.",
	}, &StaticCmd{
		CmdNames: []string{"Play", "p"},
		Desc:     "Queues up or starts playing a song, either by url or by searching what you wrote",
		LongDesc: "Queues up or starts playing a song, either by url or by searching what you wrote\nExamples:\n" +
			"`play c2c down the road` - will search for the song and play the first search result\n`play https://www.youtube.com/watch?v=k1uUIJPD0Nk` - will play the specific linked video",
	}, &StaticCmd{
		CmdNames: []string{"Queue", "q"},
		Desc:     "Shows the current queue",
	}, &StaticCmd{
		CmdNames: []string{"Skip", "s"},
		Desc:     "Skips the current video, if you're not a moderator the majority will have to vote in favor",
	}, &StaticCmd{
		CmdNames: []string{"Volume", "v"},
		Desc:     "Sets the volume, accepts a number between `1-100`",
	})

	session, err := discordgo.New(os.Getenv("DG_TOKEN"))
	if err != nil {
		log.Fatal("Failed setting up session:", err)
	}

	session.AddHandler(system.HandleMessageCreate)
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Ready recevied")
	})
	err = session.Open()
	if err != nil {
		log.Fatal("Failed opening gateway connection:", err)
	}
	log.Println("Running, Ctrl-c to stop.")
	select {}
}

type StaticCmd struct {
	Resp           string
	CmdNames       []string
	Desc, LongDesc string
	Cat            *dcmd.Category
}

// Compilie time assertions, will not compiled unless StaticCmd implements these interfaces
var _ dcmd.Cmd = (*StaticCmd)(nil)
var _ dcmd.CmdWithDescriptions = (*StaticCmd)(nil)
var _ dcmd.CmdWithCategory = (*StaticCmd)(nil)

func (s *StaticCmd) Names() []string { return s.CmdNames }

// Descriptions should return a short Desc (used in the overall help overiview) and one long descriptions for targetted help
func (s *StaticCmd) Descriptions() (string, string) { return s.Desc, "" }

func (e *StaticCmd) Run(data *dcmd.Data) (interface{}, error) {
	if e.Resp == "" {
		return "Mock response", nil
	}
	return e.Resp, nil
}

func (e *StaticCmd) Category() *dcmd.Category {
	return e.Cat
}
