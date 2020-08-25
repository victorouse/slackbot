package service

import (
	"github.com/victorouse/slackbot/bot"
	"github.com/victorouse/slackbot/config"
	"github.com/victorouse/slackbot/scheduler"
)

type Service struct {
	Bot       *bot.Bot
	Scheduler *scheduler.Scheduler
}

func NewService() (*Service, error) {
	config := config.NewConfig()

	b, err := bot.NewBot(config)
	if err != nil {
		return nil, err
	}

	s := scheduler.NewScheduler(b)

	return &Service{
		Bot:       b,
		Scheduler: s,
	}, nil
}
