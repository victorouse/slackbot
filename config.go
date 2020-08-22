package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SlackAccessToken       string `split_words:"true"`
	SlackVerificationToken string `split_words:"true"`
	Port                   string `default:"3000"`
}

var config Config

func setupConfig() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
}
