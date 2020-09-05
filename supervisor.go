package slackbot

type Supervisor struct {
	Bot  *Bot
	Cron *Cron
}

func NewSupervisor(
	bot *Bot,
	cron *Cron,
) *Supervisor {
	return &Supervisor{
		Bot:  bot,
		Cron: cron,
	}
}

func (s *Supervisor) InitActions() {
	s.Bot.Actions = map[string]*Action{
		"echo": {
			Description: "echo command arguments back",
			Run:         s.echo,
		},
		"job": {
			Description: "list|start|stop cron jobs",
			Run:         s.job,
		},
	}
}

func (s *Supervisor) InitJobs() {
	s.Cron.Jobs = map[string]*Job{
		"time": {
			Schedule:    "@every 1m",
			Description: "tells the time at an interval",
			Run:         s.tellTime,
		},
	}

	for job := range s.Cron.Jobs {
		s.Cron.StartJob(job)
	}

	s.Cron.Start()
}
