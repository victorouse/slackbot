package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

func (b *Bot) handleMessageEvent(ev *slackevents.MessageEvent) {
	if len(ev.BotID) == 0 && !strings.Contains(ev.Text, b.info.ID) && ev.BotID != b.info.Profile.BotID {
		fmt.Println("[INFO] Received message")
		b.client.PostMessage(ev.Channel, slack.MsgOptionText("Did you say something?", false))
	}
}

func (b *Bot) handleAppMentionEvent(ev *slackevents.AppMentionEvent) {
	fmt.Println("[INFO] Received slack mention")
	fmt.Printf("[INFO] Message: %s\n", ev.Text)
	parts := strings.Split(ev.Text, " ")
	command, args := parts[1], parts[2:]

	if action, ok := b.actions[command]; ok {
		if command == "help" {
			b.client.PostMessage(ev.Channel, slack.MsgOptionText(b.help(), false))
			return
		}

		result := action.run(args...)
		b.client.PostMessage(ev.Channel, slack.MsgOptionText(result, false))
	} else {
		b.client.PostMessage(ev.Channel, slack.MsgOptionText(b.help(), false))
	}
}

func (b *Bot) handleCallbackEvent(event slackevents.EventsAPIEvent) {
	innerEvent := event.InnerEvent

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		b.handleMessageEvent(ev)

	case *slackevents.AppMentionEvent:
		b.handleAppMentionEvent(ev)
	}
}

type Action struct {
	description string
	run         func(args ...string) string
}

var actions = map[string]Action{
	"echo": {
		description: "echo command arguments back",
		run: func(args ...string) string {
			return strings.Join(args, " ")
		},
	},
}

type Bot struct {
	client  *slack.Client
	auth    *slack.AuthTestResponse
	info    *slack.User
	actions map[string]Action
}

func (b *Bot) eventHandler(w http.ResponseWriter, r *http.Request) {
	event, err := parseSlackEvent(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch event.Type {
	case slackevents.URLVerification:
		handleChallengeRequest(w, r)

	case slackevents.CallbackEvent:
		b.handleCallbackEvent(event)
	}
}

func (b *Bot) help() string {
	var helpText strings.Builder

	for command, action := range b.actions {
		helpText.WriteString(fmt.Sprintf("%s\t-\t%s", command, action.description))
	}

	return "```" + helpText.String() + "```"
}

func main() {
	setupConfig()
	slackClient := slack.New(config.SlackAccessToken)

	botAuth, err := slackClient.AuthTest()
	if err != nil {
		fmt.Printf("Error authenticating: %s\n", err)
		return
	}

	botInfo, err := slackClient.GetUserInfo(botAuth.UserID)
	if err != nil {
		fmt.Printf("Error getting bot info: %s\n", err)
		return
	}

	bot := Bot{
		client:  slackClient,
		auth:    botAuth,
		info:    botInfo,
		actions: actions,
	}

	http.HandleFunc("/events", bot.eventHandler)

	fmt.Printf("[INFO] Server listening on port :%s\n", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil)
}
