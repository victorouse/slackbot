package slackbot

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

type Cron struct {
	Jobs map[string]*Job
	cron *cron.Cron
}

func NewCron() *Cron {
	cron := cron.New(cron.WithSeconds())

	return &Cron{
		Jobs: map[string]*Job{},
		cron: cron,
	}
}

func (c *Cron) AddJob(name string, schedule string, description string, cmd func()) (*Job, error) {
	fmt.Printf("[INFO] Adding job: %s\n", name)
	entryID, err := c.cron.AddFunc(schedule, cmd)
	fmt.Printf("[INFO] Entry ID: %d\n", entryID)
	if err != nil {
		return nil, err
	}

	job := &Job{
		EntryID:     entryID,
		Schedule:    schedule,
		Description: description,
		Run:         cmd,
		Active:      true,
	}

	c.Jobs[name] = job
	return job, nil
}

func (c *Cron) RemoveJob(name string) {
	fmt.Printf("[INFO] Removing job: %s\n", name)
	if job, exists := c.Jobs[name]; exists {
		delete(c.Jobs, name)
		c.cron.Remove(job.EntryID)
	}
}

func (c *Cron) StartJob(name string) error {
	if job, exists := c.Jobs[name]; exists {
		fmt.Printf("[INFO] Starting job: %s - %s\n", name, job.Schedule)
		entryID, err := c.cron.AddFunc(job.Schedule, job.Run)
		fmt.Printf("[INFO] Entry ID: %d\n", entryID)
		job.Active = true
		job.EntryID = entryID

		if err != nil {
			fmt.Printf("[ERROR] Error starting job: %s\n", name)
			return err
		}
	}

	return nil
}

func (c *Cron) UpdateSchedule(name string, schedule string) {
	fmt.Printf("[INFO] Updating job: %s\n", name)
	job := c.GetJob(name)
	if job != nil {
		c.StopJob(name)
		c.AddJob(name, schedule, job.Description, job.Run)
	}
}

func (c *Cron) StopJob(name string) {
	fmt.Printf("[INFO] Stopping job: %s\n", name)
	if job, exists := c.Jobs[name]; exists {
		c.cron.Remove(job.EntryID)
		job.Active = false
	}
}

func (c *Cron) GetJob(name string) *Job {
	if job, exists := c.Jobs[name]; exists {
		return job
	}

	return nil
}

func (c *Cron) Start() {
	go c.cron.Start()
}
