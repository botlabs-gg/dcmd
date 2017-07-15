package dcmd

import (
	"errors"
	"fmt"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/dutil"
	"reflect"
	"time"
)

type Response interface {
	// Channel, session, command etc can all be found in this context
	Send(data *Data) ([]*discordgo.Message, error)
}

func SendResponseInterface(data *Data, reply interface{}, escapeEveryoneMention bool) ([]*discordgo.Message, error) {

	switch t := reply.(type) {
	case Response:
		return t.Send(data)
	case string:
		if t != "" {
			if escapeEveryoneMention {
				t = dutil.EscapeEveryoneMention(t)
			}
			return dutil.SplitSendMessage(data.Session, data.Channel.ID, t)
		}
		return []*discordgo.Message{}, nil
	case error:
		if t != nil {
			m := t.Error()
			if escapeEveryoneMention {
				m = dutil.EscapeEveryoneMention(m)
			}
			return dutil.SplitSendMessage(data.Session, data.Channel.ID, m)
		}
		return []*discordgo.Message{}, nil
	case *discordgo.MessageEmbed:
		m, err := data.Session.ChannelMessageSendEmbed(data.Channel.ID, t)
		return []*discordgo.Message{m}, err
	case []*discordgo.MessageEmbed:
		msgs := make([]*discordgo.Message, len(t))
		for i, embed := range t {
			m, err := data.Session.ChannelMessageSendEmbed(data.Channel.ID, embed)
			if err != nil {
				return msgs, err
			}
			msgs[i] = m
		}
		return msgs, nil
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
			ids := make([]string, len(msgs))
			for i, m := range msgs {
				ids[i] = m.ID
			}
			data.Session.ChannelMessagesBulkDelete(data.Channel.ID, ids)
		} else {
			data.Session.ChannelMessageDelete(data.Channel.ID, msgs[0].ID)
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
	channelPerms, err := data.Session.State.UserChannelPermissions(data.Session.State.User.ID, data.Channel.ID)
	if err != nil {
		return nil, err
	}

	if channelPerms&discordgo.PermissionEmbedLinks != 0 {
		m, err := data.Session.ChannelMessageSendEmbed(data.Channel.ID, fe.MessageEmbed)
		if err != nil {
			return nil, err
		}

		return []*discordgo.Message{m}, nil
	}

	content := StringEmbed(fe.MessageEmbed) + "\n*I have no 'embed links' permissions here, this is a fallback. it looks prettier if i have that perm :)*"
	return dutil.SplitSendMessage(data.Session, data.Channel.ID, content)
}

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
