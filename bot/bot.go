package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/victorouse/slackbot/config"
)

type Bot struct {
	Client  *slack.Client
	Auth    *slack.AuthTestResponse
	Info    *slack.User
	Actions map[string]Action
}

func codeblock(s string) string {
	return "```" + s + "```"
}

func NewBot(config *config.Config) (*Bot, error) {
	slackClient := slack.New(config.SlackAccessToken)

	botAuth, err := slackClient.AuthTest()
	if err != nil {
		return nil, err
	}

	botInfo, err := slackClient.GetUserInfo(botAuth.UserID)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Client:  slackClient,
		Auth:    botAuth,
		Info:    botInfo,
		Actions: actions,
	}, nil
}

func (b *Bot) GetChannelIDByName(channelName string) (string, error) {
	ctx := context.Background()
	params := &slack.GetConversationsParameters{}
	channels, _, err := b.Client.GetConversationsContext(ctx, params)
	if err != nil {
		return "", err
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("No channel with name: %s\n", channelName)
}

func (b *Bot) Help() string {
	var helpText strings.Builder

	for command, action := range b.Actions {
		helpText.WriteString(fmt.Sprintf("%s - %s\n", command, action.Description))
	}

	return codeblock(helpText.String())
}
