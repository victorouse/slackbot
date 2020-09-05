package slackbot

import (
	"fmt"
	"strconv"
	"strings"
)

type Action struct {
	Description string
	Run         func(args ...string) error
}

func (s *Supervisor) echo(args ...string) error {
	message := strings.Join(args, " ")

	err := s.Bot.SendMessageToChannel("super-secret", message)
	if err != nil {
		return err
	}

	return nil
}

func (s *Supervisor) job(args ...string) error {
	command, params := args[0], args[1:]

	switch command {
	case "list":
		s.listJobs()

	case "start":
		if len(params) > 0 {
			s.startJob(params[0])
		}

	case "stop":
		if len(params) > 0 {
			s.stopJob(params[0])
		}

	case "update":
		if len(params) > 0 {
			s.updateJob(params[0], strings.Join(params[1:], " "))
		}
	}

	return nil
}

func (s *Supervisor) listJobs() {
	headers := []string{"name", "description", "schedule", "active"}
	jobs := [][]string{}
	for name, j := range s.Cron.Jobs {
		jobs = append(
			jobs,
			[]string{
				name,
				j.Description,
				j.Schedule,
				strconv.FormatBool(j.Active),
			},
		)
	}
	s.Bot.SendMessage("super-secret", Codeblock(Table(headers, jobs)))
}

func (s *Supervisor) startJob(name string) {
	if _, exists := s.Cron.Jobs[name]; exists {
		s.Cron.StartJob(name)
		s.listJobs()
	} else {
		s.Bot.SendMessage("super-secret", Codeblock(fmt.Sprintf("No job: %s", name)))
	}
}

func (s *Supervisor) stopJob(name string) {
	if _, exists := s.Cron.Jobs[name]; exists {
		s.Cron.StopJob(name)
		s.listJobs()
	} else {
		s.Bot.SendMessage("super-secret", Codeblock(fmt.Sprintf("No job: %s", name)))
	}
}

func (s *Supervisor) updateJob(name string, schedule string) {
	if _, exists := s.Cron.Jobs[name]; exists {
		s.Cron.UpdateSchedule(name, schedule)
		s.listJobs()
	} else {
		s.Bot.SendMessage("super-secret", Codeblock(fmt.Sprintf("No job: %s", name)))
	}
}
