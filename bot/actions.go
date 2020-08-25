package bot

import (
	"strings"
)

type Action struct {
	Description string
	Run         func(args ...string) string
}

var actions = map[string]Action{
	"echo": {
		Description: "echo command arguments back",
		Run:         echo,
	},
	"banger": {
		Description: "a minimum of 150bpm",
		Run:         banger,
	},
}

func banger(args ...string) string {
	return "https://www.youtube.com/watch?v=hUVxpaEcsdg"
}

func echo(args ...string) string {
	return strings.Join(args, " ")
}
