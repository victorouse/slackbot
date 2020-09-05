package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/slack-go/slack/slackevents"
	"github.com/victorouse/slackbot/config"
)

type Responder struct {
	Supervisor *Supervisor
}

func NewResponder(supervisor *Supervisor) *Responder {
	return &Responder{
		Supervisor: supervisor,
	}
}

func (s *Responder) HandleEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[INFO] Received event")
	event, err := parseSlackEvent(r)
	if err != nil {
		fmt.Println("[ERROR] Parsing event")
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

func (s *Responder) handleCallbackEvent(event slackevents.EventsAPIEvent) {
	innerEvent := event.InnerEvent

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		s.handleMessageEvent(ev)

	case *slackevents.AppMentionEvent:
		s.handleAppMentionEvent(ev)
	}
}

func (s *Responder) handleAppMentionEvent(ev *slackevents.AppMentionEvent) {
	fmt.Println("[INFO] Received slack mention")
	fmt.Printf("[INFO] Message: %s\n", ev.Text)

	parts := strings.Split(ev.Text, " ")
	command, args := parts[1], parts[2:]

	if action, exists := s.Supervisor.Bot.Actions[command]; exists {
		err := action.Run(args...)
		if err != nil {
			s.Supervisor.Bot.SendMessage(ev.Channel, "something went wrong")
		}
	} else {
		s.Supervisor.Bot.SendMessage(ev.Channel, "command not found")
	}
}

func (s *Responder) handleMessageEvent(ev *slackevents.MessageEvent) {
	fmt.Printf("[INFO] Received message from: %s\n", ev.Channel)

	shouldRespond := len(ev.BotID) == 0 &&
		ev.BotID != s.Supervisor.Bot.BotID &&
		!strings.Contains(ev.Text, s.Supervisor.Bot.UserID)

	if shouldRespond {
		fmt.Println("[INFO] Responding to message")
		s.Supervisor.Bot.SendMessage(ev.Channel, "Did you say something?")
	}
}

func parseSlackEvent(r *http.Request) (slackevents.EventsAPIEvent, error) {
	config := config.ParseConfig()
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	r.Body.Close()
	r.Body = ioutil.NopCloser(buf)

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

	err := json.Unmarshal([]byte(body), &resp)
	if err != nil {
		fmt.Println("[ERROR] Unmarshalling challenge response body")
		return nil, err
	}

	return resp, nil
}

func handleChallengeRequest(w http.ResponseWriter, r *http.Request) {
	resp, err := parseChallengeRequest(r)
	if err != nil {
		fmt.Println("[ERROR] Parsing challenge request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text")
	w.Write([]byte(resp.Challenge))
}
