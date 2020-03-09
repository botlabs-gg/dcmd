package dcmd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jonas747/discordgo"
)

type Response interface {
	// Channel, session, command etc can all be found in this context
	Send(data *Data) ([]*discordgo.Message, error)
}

func SendResponseInterface(data *Data, reply interface{}, escapeEveryoneMention bool) ([]*discordgo.Message, error) {

	allowedMentions := discordgo.AllowedMentions{}
	if !escapeEveryoneMention {
		// Legacy behaviour
		allowedMentions.Parse = []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles, discordgo.AllowedMentionTypeUsers, discordgo.AllowedMentionTypeUsers}
	}

	switch t := reply.(type) {
	case Response:
		return t.Send(data)
	case string:
		if t != "" {
			return SplitSendMessage(data.Session, data.Msg.ChannelID, t, allowedMentions)
		}
		return []*discordgo.Message{}, nil
	case error:
		if t != nil {
			m := t.Error()
			return SplitSendMessage(data.Session, data.Msg.ChannelID, m, allowedMentions)
		}
		return []*discordgo.Message{}, nil
	case *discordgo.MessageEmbed:
		m, err := data.Session.ChannelMessageSendEmbed(data.Msg.ChannelID, t)
		return []*discordgo.Message{m}, err
	case []*discordgo.MessageEmbed:
		msgs := make([]*discordgo.Message, len(t))
		for i, embed := range t {
			m, err := data.Session.ChannelMessageSendEmbed(data.Msg.ChannelID, embed)
			if err != nil {
				return msgs, err
			}
			msgs[i] = m
		}
		return msgs, nil
	case *discordgo.MessageSend:
		m, err := data.Session.ChannelMessageSendComplex(data.Msg.ChannelID, t)
		return []*discordgo.Message{m}, err
	}

	return nil, errors.New("Unknown reply type: " + reflect.TypeOf(reply).String() + " (Does not implement Response)")
}

// Temporary response deletes the inner response after Duration
type TemporaryResponse struct {
	Response       interface{}
	Duration       time.Duration
	EscapeEveryone bool
}

func NewTemporaryResponse(d time.Duration, inner interface{}, escapeEveryoneMention bool) *TemporaryResponse {
	return &TemporaryResponse{
		Duration: d, Response: inner,

		EscapeEveryone: escapeEveryoneMention,
	}
}

func (t *TemporaryResponse) Send(data *Data) ([]*discordgo.Message, error) {

	msgs, err := SendResponseInterface(data, t.Response, t.EscapeEveryone)
	if err != nil {
		return nil, err
	}

	time.AfterFunc(t.Duration, func() {
		// do a bulk if 2 or more
		if len(msgs) > 1 {
			ids := make([]int64, len(msgs))
			for i, m := range msgs {
				ids[i] = m.ID
			}
			data.Session.ChannelMessagesBulkDelete(data.Msg.ChannelID, ids)
		} else {
			data.Session.ChannelMessageDelete(data.Msg.ChannelID, msgs[0].ID)
		}
	})
	return msgs, nil
}

// The FallbackEmbed reponse type will turn the embed into a normal mesasge if there is not enough permissions
// This requires state member tracking enabled
type FallbackEmebd struct {
	*discordgo.MessageEmbed
}

func (fe *FallbackEmebd) Send(data *Data) ([]*discordgo.Message, error) {
	channelPerms, err := data.Session.State.UserChannelPermissions(data.Session.State.User.ID, data.Msg.ChannelID)
	if err != nil {
		return nil, err
	}

	if channelPerms&discordgo.PermissionEmbedLinks != 0 {
		m, err := data.Session.ChannelMessageSendEmbed(data.Msg.ChannelID, fe.MessageEmbed)
		if err != nil {
			return nil, err
		}

		return []*discordgo.Message{m}, nil
	}

	content := StringEmbed(fe.MessageEmbed) + "\n*I have no 'embed links' permissions here, this is a fallback. it looks prettier if i have that perm :)*"
	return SplitSendMessage(data.Session, data.Msg.ChannelID, content, discordgo.AllowedMentions{})
}

// StringEmbed turns the embed into the best
func StringEmbed(embed *discordgo.MessageEmbed) string {
	body := ""

	if embed.Author != nil {
		body += embed.Author.Name + "\n"
		body += embed.Author.URL + "\n"
	}

	if embed.Title != "" {
		body += "**" + embed.Title + "**\n"
	}

	if embed.Description != "" {
		body += embed.Description + "\n"
	}
	if body != "" {
		body += "\n"
	}

	for _, v := range embed.Fields {
		body += fmt.Sprintf("**%s**\n%s\n\n", v.Name, v.Value)
	}
	return body
}

// SplitSendMessage uses SplitString to make sure each message is within 2k characters and splits at last newline before that (if possible)
func SplitSendMessage(s *discordgo.Session, channelID int64, contents string, allowedMentions discordgo.AllowedMentions) ([]*discordgo.Message, error) {
	result := make([]*discordgo.Message, 0, 1)

	split := SplitString(contents, 2000)
	for _, v := range split {
		m, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content:         v,
			AllowedMentions: allowedMentions,
		})
		if err != nil {
			return result, err
		}

		result = append(result, m)
	}

	return result, nil
}

// SplitString uses StrSplitNext to split a string at the last newline before maxLen, throwing away leading and ending whitespaces in the process
func SplitString(s string, maxLen int) []string {
	result := make([]string, 0, 1)

	rest := s
	for {
		if strings.TrimSpace(rest) == "" {
			break
		}

		var split string
		split, rest = StrSplitNext(rest, maxLen)

		split = strings.TrimSpace(split)
		if split == "" {
			continue
		}

		result = append(result, split)
	}

	return result
}

// StrSplitNext Will split "s" before runecount at last possible newline, whitespace or just at "runecount" if there is no whitespace
// If the runecount in "s" is less than "runeCount" then "last" will be zero
func StrSplitNext(s string, runeCount int) (split, rest string) {
	if utf8.RuneCountInString(s) <= runeCount {
		return s, ""
	}

	_, beforeIndex := RuneByIndex(s, runeCount)
	firstPart := s[:beforeIndex]

	// Split at newline if possible
	foundWhiteSpace := false
	lastIndex := strings.LastIndex(firstPart, "\n")
	if lastIndex == -1 {
		// No newline, check for any possible whitespace then
		lastIndex = strings.LastIndexFunc(firstPart, func(r rune) bool {
			return unicode.In(r, unicode.White_Space)
		})
		if lastIndex == -1 {
			lastIndex = beforeIndex
		} else {
			foundWhiteSpace = true
		}
	} else {
		foundWhiteSpace = true
	}

	// Remove the whitespace we split at if any
	if foundWhiteSpace {
		_, rLen := utf8.DecodeRuneInString(s[lastIndex:])
		rest = s[lastIndex+rLen:]
	} else {
		rest = s[lastIndex:]
	}

	split = s[:lastIndex]

	return
}

// RuneByIndex Returns the string index from the rune position
// Panics if utf8.RuneCountInString(s) <= runeIndex or runePos < 0
func RuneByIndex(s string, runePos int) (rune, int) {
	sLen := utf8.RuneCountInString(s)
	if sLen <= runePos || runePos < 0 {
		panic("runePos is out of bounds")
	}

	i := 0
	last := rune(0)
	for k, r := range s {
		if i == runePos {
			return r, k
		}
		i++
		last = r
	}
	return last, i
}
