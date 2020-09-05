package slackbot

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/victorouse/slackbot/config"
)

type Bot struct {
	Client  *slack.Client
	BotID   string
	UserID  string
	Actions map[string]*Action
}

func NewBot() (*Bot, error) {
	config := config.ParseConfig()
	client := slack.New(config.SlackAccessToken)

	auth, err := client.AuthTest()
	if err != nil {
		return nil, err
	}

	info, err := client.GetUserInfo(auth.UserID)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Client: client,
		UserID: auth.UserID,
		BotID:  info.Profile.BotID,
	}, nil
}

func (b *Bot) SendMessage(channelId string, message string) error {
	_, _, err := b.Client.PostMessage(channelId, slack.MsgOptionText(message, false))
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) SendMessageToChannel(channelName string, message string) error {
	channel, err := b.GetChannelIDByName(channelName)
	if err != nil {
		return err
	}

	return b.SendMessage(channel, message)
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
