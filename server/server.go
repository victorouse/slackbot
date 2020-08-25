package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/victorouse/slackbot/bot"
	"github.com/victorouse/slackbot/config"
	"github.com/victorouse/slackbot/scheduler"
)

func parseSlackEvent(r *http.Request) (slackevents.EventsAPIEvent, error) {
	config := config.NewConfig()
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

func parseChallengeRequest(r *http.Request) (*slackevents.ChallengeResponse, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	var resp *slackevents.ChallengeResponse

	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func handleChallengeRequest(w http.ResponseWriter, r *http.Request) {
	resp, err := parseChallengeRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text")
	w.Write([]byte(resp.Challenge))
}

type Server struct {
	httpServer *http.Server
	Bot        *bot.Bot
	Scheduler  *scheduler.Scheduler
}

func NewServer() (*Server, error) {
	config := config.NewConfig()
	server := &http.Server{
		Addr: config.Port,
	}

	b, err := bot.NewBot(config)
	if err != nil {
		return nil, err
	}

	s := scheduler.NewScheduler()

	return &Server{
		httpServer: server,
		Bot:        b,
		Scheduler:  s,
	}, nil
}

func (s *Server) ListenAndServe() {
	config := config.NewConfig()
	port := fmt.Sprintf(":%s", config.Port)
	s.httpServer.ListenAndServe(port)
}

func (s *Server) HandleEvent(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleCallbackEvent(event slackevents.EventsAPIEvent) {
	innerEvent := event.InnerEvent

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		s.handleMessageEvent(ev)

	case *slackevents.AppMentionEvent:
		s.handleAppMentionEvent(ev)
	}
}

func (s *Server) handleAppMentionEvent(ev *slackevents.AppMentionEvent) {
	fmt.Println("[INFO] Received slack mention")
	fmt.Printf("[INFO] Message: %s\n", ev.Text)
	parts := strings.Split(ev.Text, " ")
	command, args := parts[1], parts[2:]

	if action, ok := s.bot.Actions[command]; ok {
		if command == "help" {
			s.bot.Client.PostMessage(ev.Channel, slack.MsgOptionText(s.bot.Help(), false))
			return
		}

		result := action.Run(args...)
		s.bot.Client.PostMessage(ev.Channel, slack.MsgOptionText(result, false))
	} else {
		s.bot.Client.PostMessage(ev.Channel, slack.MsgOptionText(s.bot.Help(), false))
	}
}

func (s *Server) handleMessageEvent(ev *slackevents.MessageEvent) {
	if len(ev.BotID) == 0 && !strings.Contains(ev.Text, s.bot.Info.ID) && ev.BotID != s.bot.Info.Profile.BotID {
		fmt.Println("[INFO] Received message")
		s.bot.Client.PostMessage(ev.Channel, slack.MsgOptionText("Did you say something?", false))
	}
}
