package dcmd

import (
	"fmt"
	"github.com/jonas747/discordgo"
	"github.com/pkg/errors"
	"log"
	"runtime/debug"
	"strings"
)

type System struct {
	Root           *Container
	Prefix         PrefixProvider
	ResponseSender ResponseSender
}

func NewStandardSystem(staticPrefix string) (system *System) {
	sys := &System{
		Root:           &Container{HelpTitleEmoji: "ℹ️", HelpColor: 0xbeff7a},
		ResponseSender: &StdResponseSender{LogErrors: true},
	}
	if staticPrefix != "" {
		sys.Prefix = NewSimplePrefixProvider(staticPrefix)
	}

	sys.Root.AddMidlewares(ArgParserMW)

	return sys
}

// You can add this as a handler directly to discordgo, it will recover from any panics that occured in commands
// and log errors using the standard logger
func (sys *System) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Set up handler to recover from panics
	defer func() {
		if r := recover(); r != nil {
			sys.handlePanic(s, r, false)
		}
	}()

	err := sys.CheckMessage(s, m)
	if err != nil {
		log.Println("[DCMD ERROR]: Failed checking message:", err)
	}
}

// CheckMessage checks the message for commands, and triggers any command that the message should trigger
// you should not add this as an discord handler directly, if you want to do that you should add "system.HandleMessageCreate" instead.
func (sys *System) CheckMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	data, err := sys.FillData(s, m.Message)
	if err != nil {
		return err
	}

	if !sys.FindPrefix(data) {
		// No prefix found in the message for a command to be triggered
		return nil
	}

	response, err := sys.Root.Run(data)
	return sys.ResponseSender.SendResponse(data, response, err)
}

// FindPrefix checks if the message has a proper command prefix (either from the PrefixProvider or a direction mention to the bot)
// It sets the source field, and MsgStripped in data if found
func (sys *System) FindPrefix(data *Data) (found bool) {
	if data.Channel.IsPrivate {
		data.MsgStrippedPrefix = data.Msg.Content
		data.Source = DMSource
		return true
	}

	if sys.FindMentionPrefix(data) {
		return true
	}

	// Check for custom prefix
	if sys.Prefix == nil {
		return false
	}

	prefix := sys.Prefix.Prefix(data)
	if prefix == "" {
		return false
	}

	data.PrefixUsed = prefix

	if strings.HasPrefix(data.Msg.Content, prefix) {
		data.Source = PrefixSource
		data.MsgStrippedPrefix = strings.TrimSpace(strings.Replace(data.Msg.Content, prefix, "", 1))
		found = true
	}

	return
}

func (sys *System) FindMentionPrefix(data *Data) (found bool) {
	if data.Session.State.User == nil {
		return false
	}

	ok := false
	stripped := ""

	// Check for mention
	id := data.Session.State.User.ID
	if strings.Index(data.Msg.Content, "<@"+id+">") == 0 { // Normal mention
		ok = true
		stripped = strings.Replace(data.Msg.Content, "<@"+id+">", "", 1)
		data.PrefixUsed = "<@" + id + ">"
	} else if strings.Index(data.Msg.Content, "<@!"+id+">") == 0 { // Nickname mention
		ok = true
		data.PrefixUsed = "<@!" + id + ">"
		stripped = strings.Replace(data.Msg.Content, "<@!"+id+">", "", 1)
	}

	if ok {
		data.MsgStrippedPrefix = strings.TrimSpace(stripped)
		data.Source = MentionSource

		return true
	}

	return false

}

func (sys *System) FillData(s *discordgo.Session, m *discordgo.Message) (*Data, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return nil, errors.Wrap(err, "System.FillData")
	}

	data := &Data{
		Msg:     m,
		Channel: channel,
		Session: s,
		System:  sys,
	}

	if !channel.IsPrivate {
		g, err := s.State.Guild(channel.GuildID)
		if err != nil {
			return nil, errors.Wrap(err, "System.FillData")
		}

		data.Guild = g
	} else {
		data.Source = DMSource
	}

	return data, nil
}

func (sys *System) handlePanic(s *discordgo.Session, r interface{}, sendChatNotice bool) {
	// TODO
	stack := debug.Stack()
	log.Printf("[DCMD PANIC] %v\n%S", r, string(stack))
}

// Retrieves the prefix that might be different on a per server basis
type PrefixProvider interface {
	Prefix(data *Data) string
}

// Simple Prefix provider for global fixed prefixes
type SimplePrefixProvider struct {
	prefix string
}

func NewSimplePrefixProvider(prefix string) PrefixProvider {
	return &SimplePrefixProvider{prefix: prefix}
}

func (pp *SimplePrefixProvider) Prefix(d *Data) string {
	return pp.prefix
}

func Indent(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "-"
	}
	return indent
}

type ResponseSender interface {
	SendResponse(cmdData *Data, resp interface{}, err error) error
}

type StdResponseSender struct {
	LogErrors bool
}

func (s *StdResponseSender) SendResponse(cmdData *Data, resp interface{}, err error) error {
	if err != nil && s.LogErrors {
		log.Printf("[DCMD]: Command %q returned an error: %s", cmdData.Cmd.FormatNames(false, "/"), err)
	}

	var errR error
	if resp == nil && err != nil {
		_, errR = SendResponseInterface(cmdData, fmt.Sprintf("%q command returned an error: %s", cmdData.Cmd.FormatNames(false, "/"), err), true)
	} else if resp != nil {
		_, errR = SendResponseInterface(cmdData, resp, false)
	}

	return errR
}
