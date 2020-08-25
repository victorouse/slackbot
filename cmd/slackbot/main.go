package main

import (
	"log"

	"github.com/victorouse/slackbot/server"
)

func main() {
	s, err := server.NewServer()

	if err != nil {
		log.Fatal("Could not start server")
	}

	s.HttpServer.ListenAndServe()
}
