package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/victorouse/slackbot/bot"
)

type Job struct {
	entryID     cron.EntryID
	schedule    string
	description string
	active      bool
	cmd         func(bot *bot.Bot)
}

var jobs = []Job{
	{
		schedule:    "@every 10s",
		description: "tells the time",
		cmd:         tellTime,
	},
}

func tellTime(b *bot.Bot) {
	channelID, err := b.GetChannelIDByName("super-secret")
	if err != nil {
		fmt.Printf("Error getting channel info: %s\n", err)
		return
	}

	now := time.Now().Format(time.UnixDate)
	b.Client.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf(":clock1: The time is now: %s\n", now), false))
}
