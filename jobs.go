package slackbot

import (
	"fmt"
	"time"

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
	t := time.Now().Format(time.UnixDate)
	message := fmt.Sprintf(":clock1: The time is now: %s :clock1:\n", t)
	s.Bot.SendMessageToChannel("super-secret", message)
}

func (s *Supervisor) sendSOTD() {
	if s.DAO.Store.sotd != "" {
		s.Bot.SendMessage("super-secret", s.DAO.Store.sotd)
	} else {
		s.Bot.SendMessage("super-secret", "No song of the day set :(")
	}
}
