package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SlackAccessToken       string `split_words:"true"`
	SlackVerificationToken string `split_words:"true"`
	Port                   string `default:"3000"`
}

func ParseConfig() *Config {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
		panic(1)
	}

	return &config
}
