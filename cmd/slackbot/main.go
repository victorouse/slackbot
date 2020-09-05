package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/victorouse/slackbot"
)

func main() {
	bot, err := slackbot.NewBot()
	if err != nil {
		log.Fatal("Could not initialise bot")
		return
	}
	cron := slackbot.NewCron()
	supervisor := slackbot.NewSupervisor(bot, cron)
	responder := slackbot.NewResponder(supervisor)

	supervisor.InitActions()
	supervisor.InitJobs()

	http.HandleFunc("/events", responder.HandleEvent)

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
