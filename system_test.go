package dcmd

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TestUserID = "105487308693757952"
)

var (
	testSystem  *System
	testSession *discordgo.Session
)

type TestCommand struct{}

func (e *TestCommand) Names() []string          { return []string{"test"} }
func (e *TestCommand) ShortDescription() string { return "Test Description" }
func (e *TestCommand) Run(data *Data) (interface{}, error) {
	return "Test Response", nil
}

func SetupTestSystem() {
	testSystem, _ = NewStandardSystem("!")
	testSystem.Root.AddCommand(&TestCommand{})

	testSession = &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				User: &discordgo.User{
					ID: TestUserID,
				},
			},
		},
	}

}

func TestFindPrefix(t *testing.T) {
	testChannelNoPriv := &discordgo.Channel{
		IsPrivate: false,
	}

	testChannelPriv := &discordgo.Channel{
		IsPrivate: true,
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
		{testChannelNoPriv, "<@" + TestUserID + ">cmd", "cmd", true, MentionSource, []*discordgo.User{&discordgo.User{ID: TestUserID}}},
		{testChannelNoPriv, "<@" + TestUserID + "> cmd", "cmd", true, MentionSource, []*discordgo.User{&discordgo.User{ID: TestUserID}}},
		{testChannelNoPriv, "<@" + TestUserID + " cmd", "", false, MentionSource, nil},
		{testChannelPriv, "cmd", "cmd", true, DMSource, nil},
	}

	for k, v := range cases {
		t.Run(fmt.Sprintf("#%d-p:%v-m:%v", k, v.channel == testChannelPriv, v.shouldBeFound), func(t *testing.T) {
			testData := &Data{
				Session: testSession,
				Channel: v.channel,
				Msg: &discordgo.Message{
					Content:  v.msgContent,
					Mentions: v.mentions,
				},
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
