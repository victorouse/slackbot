package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func parseSlackEvent(r *http.Request) (slackevents.EventsAPIEvent, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	return slackevents.ParseEvent(
		json.RawMessage(body),
		slackevents.OptionVerifyToken(&slackevents.TokenComparator{
			VerificationToken: config.SlackVerificationToken,
		}),
	)
}

func handleChallengeRequest(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	var resp *slackevents.ChallengeResponse

	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text")
	w.Write([]byte(resp.Challenge))
}

type Action struct {
	description string
	run         func(svc *Service, args ...string) string
}

type Bot struct {
	client  *slack.Client
	auth    *slack.AuthTestResponse
	info    *slack.User
	actions map[string]Action
}

func NewBot(config Config) (*Bot, error) {
	slackClient := slack.New(config.SlackAccessToken)

	botAuth, err := slackClient.AuthTest()
	if err != nil {
		fmt.Printf("Error authenticating: %s\n", err)
		return &Bot{}, err
	}

	botInfo, err := slackClient.GetUserInfo(botAuth.UserID)
	if err != nil {
		fmt.Printf("Error getting bot info: %s\n", err)
		return &Bot{}, err
	}

	return &Bot{
		client:  slackClient,
		auth:    botAuth,
		info:    botInfo,
		actions: actions,
	}, nil
}

func (b *Bot) getChannelIDByName(channelName string) (string, error) {
	ctx := context.Background()
	params := &slack.GetConversationsParameters{}
	channels, _, err := b.client.GetConversationsContext(ctx, params)
	if err != nil {
		fmt.Printf("Error getting channels: %s\n", err)
		return "", err
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("No channel with name: %s\n", channelName)
}

func (b *Bot) help() string {
	var helpText strings.Builder

	for command, action := range b.actions {
		helpText.WriteString(fmt.Sprintf("%s - %s\n", command, action.description))
	}

	return "```" + helpText.String() + "```"
}

type Job struct {
	entryID     cron.EntryID
	schedule    string
	description string
	active      bool
	cmd         func(svc *Service)
}

type Scheduler struct {
	cron *cron.Cron
	jobs []Job
}

func NewScheduler() *Scheduler {
	cron := cron.New()

	return &Scheduler{
		cron: cron,
		jobs: []Job{},
	}
}

type Service struct {
	bot       *Bot
	scheduler *Scheduler
}

func (s *Service) Start() {
	for _, job := range jobs {
		s.AddJob(job)
	}

	go s.scheduler.cron.Run()
}

func (s *Service) eventHandler(w http.ResponseWriter, r *http.Request) {
	event, err := parseSlackEvent(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch event.Type {
	case slackevents.URLVerification:
		handleChallengeRequest(w, r)

	case slackevents.CallbackEvent:
		s.handleCallbackEvent(event)
	}
}

func (s *Service) handleCallbackEvent(event slackevents.EventsAPIEvent) {
	innerEvent := event.InnerEvent

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		s.handleMessageEvent(ev)

	case *slackevents.AppMentionEvent:
		s.handleAppMentionEvent(ev)
	}
}

func (s *Service) handleMessageEvent(ev *slackevents.MessageEvent) {
	if len(ev.BotID) == 0 && !strings.Contains(ev.Text, s.bot.info.ID) && ev.BotID != s.bot.info.Profile.BotID {
		fmt.Println("[INFO] Received message")
		s.bot.client.PostMessage(ev.Channel, slack.MsgOptionText("Did you say something?", false))
	}
}

func (s *Service) handleAppMentionEvent(ev *slackevents.AppMentionEvent) {
	fmt.Println("[INFO] Received slack mention")
	fmt.Printf("[INFO] Message: %s\n", ev.Text)
	parts := strings.Split(ev.Text, " ")
	command, args := parts[1], parts[2:]

	if action, ok := s.bot.actions[command]; ok {
		if command == "help" {
			s.bot.client.PostMessage(ev.Channel, slack.MsgOptionText(s.bot.help(), false))
			return
		}

		result := action.run(s, args...)
		s.bot.client.PostMessage(ev.Channel, slack.MsgOptionText(result, false))
	} else {
		s.bot.client.PostMessage(ev.Channel, slack.MsgOptionText(s.bot.help(), false))
	}
}

func (s *Service) ListJobs() string {
	var jobList strings.Builder

	for _, job := range s.scheduler.jobs {
		status := "stopped"
		if job.active {
			status = "active"
		}

		jobList.WriteString(fmt.Sprintf("[ID: %d] - %s - %s [%s]\n", job.entryID, job.schedule, job.description, status))
	}

	return "```" + jobList.String() + "```"
}

func (s *Service) GetJob(entryID cron.EntryID) (Job, bool) {
	for _, job := range s.scheduler.jobs {
		if job.entryID == entryID {
			return job, true
		}
	}

	return Job{}, false
}

func (s *Service) RemoveJob(entryID cron.EntryID) {
	newJobs := []Job{}
	for _, job := range s.scheduler.jobs {
		if job.entryID != entryID {
		}
	}

	s.scheduler.jobs = newJobs
}

func (s *Service) AddJob(job Job) error {
	entryID, err := s.scheduler.cron.AddFunc(job.schedule, func() { job.cmd(s) })
	if err != nil {
		return err
	}

	job.active = true
	job.entryID = entryID
	s.scheduler.jobs = append(s.scheduler.jobs, job)

	return nil
}

func (s *Service) StartJob(entryID cron.EntryID) string {
	for i, job := range s.scheduler.jobs {
		if job.entryID == entryID {
			id, err := s.scheduler.cron.AddFunc(job.schedule, func() { job.cmd(s) })
			if err != nil {
				return "```could not start job```"
			}
			job.entryID = id
			job.active = true
			s.scheduler.jobs[i] = job
		}
	}

	return s.ListJobs()
}

func (s *Service) StopJob(entryID cron.EntryID) string {
	for i, job := range s.scheduler.jobs {
		if job.entryID == entryID {
			s.scheduler.cron.Remove(entryID)
			job.active = false
			s.scheduler.jobs[i] = job
		}
	}

	return s.ListJobs()
}

var jobs = []Job{
	{
		schedule:    "@every 10s",
		description: "tells the time",
		cmd: func(s *Service) {
			channelID, err := s.bot.getChannelIDByName("super-secret")
			if err != nil {
				fmt.Printf("Error getting bot info: %s\n", err)
				return
			}

			now := time.Now().Format(time.UnixDate)
			s.bot.client.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf(":clock1: The time is now: %s\n", now), false))
		},
	},
}

var actions = map[string]Action{
	"echo": {
		description: "echo command arguments back",
		run: func(svc *Service, args ...string) string {
			return strings.Join(args, " ")
		},
	},
	"cron": {
		description: "cron [list|start <id>|stop <id>|set <id> <schedule>]",
		run: func(svc *Service, args ...string) string {
			if len(args) > 0 {
				subCmd := args[0]
				switch subCmd {
				case "list":
					return svc.ListJobs()

				case "start":
					if len(args) != 2 {
						return "```missing <id>```"
					}

					id, err := strconv.Atoi(args[1])
					if err != nil {
						return "```<id> is invalid```"
					}

					return svc.StartJob(cron.EntryID(id))

				case "stop":
					if len(args) != 2 {
						return "```missing <id>```"
					}

					id, err := strconv.Atoi(args[1])
					if err != nil {
						return "```<id> is invalid```"
					}

					return svc.StopJob(cron.EntryID(id))

				case "set":
					if len(args) < 3 {
						return "```missing <id> and/or <schedule>```"
					}

					id, err := strconv.Atoi(args[1])
					if err != nil {
						return "```<id> is invalid```"
					}

					job, exists := svc.GetJob(cron.EntryID(id))
					if !exists {
						return "```job not found```"
					}

					spec := strings.Join(args[2:], " ")
					job.schedule = spec

					svc.StopJob(job.entryID)
					svc.RemoveJob(job.entryID)

					err = svc.AddJob(job)
					if err != nil {
						return "```could not update job```"
					}

					return svc.ListJobs()
				}
			}

			return svc.ListJobs()
		},
	},
	"banger": {
		description: "a minimum of 150bpm",
		run: func(svc *Service, args ...string) string {
			return "https://www.youtube.com/watch?v=hUVxpaEcsdg"
		},
	},
}

func main() {
	config := NewConfig()
	bot, err := NewBot(config)
	if err != nil {
		fmt.Printf("Error creating bot: %s\n", err)
		return
	}
	scheduler := NewScheduler()
	svc := Service{bot, scheduler}
	svc.Start()

	http.HandleFunc("/events", svc.eventHandler)
	fmt.Printf("[INFO] Server listening on port :%s\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil)
}
