package scheduler

import (
	"log"
	"time"
)

// Runner is anything that can be executed on a schedule.
type Runner interface {
	Run() error
}

// Scheduler repeatedly invokes a Runner at a fixed interval.
type Scheduler struct {
	runner   Runner
	interval time.Duration
	logger   *log.Logger
}

// New creates a new Scheduler.
func New(runner Runner, interval time.Duration, logger *log.Logger) *Scheduler {
	return &Scheduler{
		runner:   runner,
		interval: interval,
		logger:   logger,
	}
}

// Start begins the scheduling loop and blocks until the done channel is closed.
func (s *Scheduler) Start(done <-chan struct{}) {
	s.logger.Printf("scheduler: starting with interval %s", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start.
	if err := s.runner.Run(); err != nil {
		s.logger.Printf("scheduler: run error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.runner.Run(); err != nil {
				s.logger.Printf("scheduler: run error: %v", err)
			}
		case <-done:
			s.logger.Println("scheduler: stopping")
			return
		}
	}
}
