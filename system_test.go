package dcmd

import (
	"fmt"
	"testing"

	"github.com/jonas747/discordgo"
	"github.com/stretchr/testify/assert"
)

const (
	TestUserID    = 105487308693757952
	TestUserIDStr = "105487308693757952"
)

var (
	testSystem  *System
	testSession *discordgo.Session
)

type TestCommand struct{}

const (
	TestResponse = "Test Response"
)

func (e *TestCommand) ShortDescription() string { return "Test Description" }
func (e *TestCommand) Run(data *Data) (interface{}, error) {
	return TestResponse, nil
}

func SetupTestSystem() {
	testSystem = NewStandardSystem("!")
	testSystem.Root.AddCommand(&TestCommand{}, NewTrigger("test"))

	testSession = &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				User: &discordgo.SelfUser{
					User: &discordgo.User{
						ID: TestUserID,
					},
				},
			},
		},
	}

}

func TestFindPrefix(t *testing.T) {
	testChannelNoPriv := &discordgo.Channel{
		Type: discordgo.ChannelTypeGuildText,
	}

	testChannelPriv := &discordgo.Channel{
		Type: discordgo.ChannelTypeDM,
	}

	cases := []struct {
		channel          *discordgo.Channel
		msgContent       string
		expectedStripped string
		shouldBeFound    bool
		expectedSource   TriggerSource
		mentions         []*discordgo.User
	}{
		{testChannelNoPriv, "!cmd", "cmd", true, PrefixSource, nil},
		{testChannelNoPriv, "cmd", "cmd", false, PrefixSource, nil},
		{testChannelNoPriv, "<@" + TestUserIDStr + ">cmd", "cmd", true, MentionSource, []*discordgo.User{&discordgo.User{ID: TestUserID}}},
		{testChannelNoPriv, "<@" + TestUserIDStr + "> cmd", "cmd", true, MentionSource, []*discordgo.User{&discordgo.User{ID: TestUserID}}},
		{testChannelNoPriv, "<@" + TestUserIDStr + " cmd", "", false, MentionSource, nil},
		{testChannelPriv, "cmd", "cmd", true, DMSource, nil},
	}

	for k, v := range cases {
		t.Run(fmt.Sprintf("#%d-p:%v-m:%v", k, v.channel == testChannelPriv, v.shouldBeFound), func(t *testing.T) {
			testData := &Data{
				Session: testSession,
				// Channel: v.channel,
				Msg: &discordgo.Message{
					Content:  v.msgContent,
					Mentions: v.mentions,
				},
			}

			if v.expectedSource != DMSource {
				testData.Msg.GuildID = 1
			}

			found := testSystem.FindPrefix(testData)
			assert.Equal(t, v.shouldBeFound, found, "Should match test case")
			if !found {
				return
			}
			assert.Equal(t, v.expectedStripped, testData.MsgStrippedPrefix, "Should be stripped off of prefix correctly")
			assert.Equal(t, v.expectedSource, testData.Source, "Should have the proper prefix")
		})
	}
}
