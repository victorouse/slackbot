package slackbot

type Supervisor struct {
	Bot  *Bot
	Cron *Cron
	DAO  *DAO
}

func NewSupervisor(
	bot *Bot,
	cron *Cron,
	dao *DAO,
) *Supervisor {
	return &Supervisor{
		Bot:  bot,
		Cron: cron,
		DAO:  dao,
	}
}

func (s *Supervisor) InitActions() {
	s.Bot.Actions = map[string]*Action{
		"echo": {
			Description: "echo command arguments back",
			Run:         s.echo,
		},
		"job": {
			Description: "list|start|stop|update cron jobs",
			Run:         s.job,
		},
		"query": {
			Description: "to be or not to be",
			Run:         s.query,
		},
		"sotd": {
			Description: "set the song of the day",
			Run:         s.sotd,
		},
		"help": {
			Description: "get help",
			Run:         s.help,
		},
	}
}

func (s *Supervisor) InitJobs() {
	s.Cron.Jobs = map[string]*Job{
		"time": {
			Schedule:    "@every 1h",
			Description: "tells the time at an interval",
			Run:         s.tellTime,
		},
		"sotd": {
			Schedule:    "* 0 9 * * *",
			Description: "the song of the day",
			Run:         s.sendSOTD,
		},
	}

	for job := range s.Cron.Jobs {
		s.Cron.StartJob(job)
	}

	s.Cron.Start()
}
