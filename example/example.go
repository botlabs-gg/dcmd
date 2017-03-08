package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dcmd"
	"log"
	"os"
)

func main() {
	system := dcmd.NewStandardSystem("[")
	system.Root.AddCommands(&EchoCmd{}, &CleanCmd{})
	system.Root.AddMidlewares(dcmd.ArgParserMW)

	session, err := discordgo.New(os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal("Failed setting up session:", err)
	}

	session.AddHandler(system.HandleMessageCreate)

	err = session.Open()
	if err != nil {
		log.Fatal("Failed opening gateway connection:", err)
	}
	log.Println("Running")
	select {}
}

type EchoCmd struct{}

func (e *EchoCmd) Names() []string          { return []string{"echo"} }
func (e *EchoCmd) ShortDescription() string { return "Echoes back what you said" }

func (e *EchoCmd) ArgDefs() (args []*dcmd.ArgDef, required int, combos [][]int) {
	return []*dcmd.ArgDef{
		{Name: "Str", Help: "What to echo back", Type: dcmd.String},
	}, 1, nil
}

func (e *EchoCmd) Run(data *dcmd.Data) (interface{}, error) {
	return data.Args[0].Value, nil
}

type CleanCmd struct{}

func (e *CleanCmd) Names() []string          { return []string{"clean"} }
func (e *CleanCmd) ShortDescription() string { return "Cleans shit up" }
func (e *CleanCmd) Switches() []*dcmd.ArgDef {
	return []*dcmd.ArgDef{
		{Name: "u", Help: "User", Type: dcmd.User},
	}
}

func (e *CleanCmd) ArgDefs() (args []*dcmd.ArgDef, required int, combos [][]int) {
	return []*dcmd.ArgDef{
		{Name: "Num", Help: "Number of messages to delete", Type: &dcmd.IntArg{Min: 1, Max: 100}},
	}, 1, nil
}

func (e *CleanCmd) Run(data *dcmd.Data) (interface{}, error) {
	return fmt.Sprintf("Should clean %d messages. Are we filtering against a user? %v", data.Args[0].Value, data.Switch("u").Value != nil), nil
}
