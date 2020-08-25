package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/victorouse/slackbot/server"
)

type Job struct {
	entryID     cron.EntryID
	schedule    string
	description string
	active      bool
	cmd         func(srv *server.Server)
}

var jobs = []Job{
	{
		schedule:    "@every 10s",
		description: "tells the time",
		cmd:         tellTime,
	},
}

func tellTime(s *server.Server) {
	channelID, err := s.Bot.GetChannelIDByName("super-secret")
	if err != nil {
		fmt.Errorf("Error getting channel info: %s\n", err)
		return
	}

	now := time.Now().Format(time.UnixDate)
	s.Bot.Client.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf(":clock1: The time is now: %s\n", now), false))
}
