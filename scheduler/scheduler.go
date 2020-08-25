package scheduler

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	jobs []Job
}

func NewScheduler() *Scheduler {
	cron := cron.New()

	return &Scheduler{
		cron: cron,
		jobs: []Job{},
	}
}

func (s *Scheduler) Start() {
	for _, job := range s.jobs {
		s.AddJob(job)
	}
}

func (s *Scheduler) PrintJobs() string {
	var jobList strings.Builder

	for _, job := range s.jobs {
		status := "stopped"
		if job.active {
			status = "active"
		}

		jobList.WriteString(fmt.Sprintf("[ID: %d] - %s - %s [%s]\n", job.entryID, job.schedule, job.description, status))
	}

	return jobList.String()
}

func (s *Scheduler) GetJob(entryID cron.EntryID) (Job, bool) {
	for _, job := range s.jobs {
		if job.entryID == entryID {
			return job, true
		}
	}

	return Job{}, false
}

func (s *Scheduler) RemoveJob(entryID cron.EntryID) {
	newJobs := []Job{}
	for _, job := range s.jobs {
		if job.entryID != entryID {
			s.jobs = append(s.jobs, job)
		}
	}

	// TODO: Remove job from cron?
	s.jobs = newJobs
}

func (s *Scheduler) AddJob(job Job) error {
	// TODO: figure out how to pass service/bot
	entryID, err := s.cron.AddFunc(job.schedule, func() { job.cmd(s) })
	if err != nil {
		return err
	}

	job.active = true
	job.entryID = entryID
	s.jobs = append(s.jobs, job)

	return nil
}

func (s *Scheduler) StartJob(entryID cron.EntryID) (bool, error) {
	started := false

	for i, job := range s.jobs {
		if job.entryID == entryID {
			// TODO: figure out how to pass service/bot
			id, err := s.cron.AddFunc(job.schedule, func() { job.cmd(s) })
			if err != nil {
				return false, err
			}

			job.entryID = id
			job.active = true
			s.jobs[i] = job
			started = true
		}
	}

	return started, nil
}

func (s *Scheduler) StopJob(entryID cron.EntryID) bool {
	found := false

	for i, job := range s.jobs {
		if job.entryID == entryID {
			s.cron.Remove(entryID)
			job.active = false
			s.jobs[i] = job
			found = true
		}
	}

	return found

}
