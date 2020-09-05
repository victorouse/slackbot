package slackbot

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

type Job struct {
	EntryID     cron.EntryID
	Schedule    string
	Description string
	Run         func()
	Active      bool
}

func (s *Supervisor) tellTime() {
	fmt.Println("tellTime")
	s.Bot.SendMessageToChannel("super-secret", "hello")
}
