package scheduler_test

import (
	"errors"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/scheduler"
)

type countingRunner struct {
	count int64
	err   error
}

func (r *countingRunner) Run() error {
	atomic.AddInt64(&r.count, 1)
	return r.err
}

func testLogger() *log.Logger {
	return log.New(os.Discard, "", 0)
}

func TestScheduler_RunsImmediatelyOnStart(t *testing.T) {
	runner := &countingRunner{}
	s := scheduler.New(runner, 10*time.Second, testLogger())
	done := make(chan struct{})

	go s.Start(done)
	time.Sleep(50 * time.Millisecond)
	close(done)
	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt64(&runner.count) < 1 {
		t.Error("expected runner to be called at least once immediately")
	}
}

func TestScheduler_RunsOnTick(t *testing.T) {
	runner := &countingRunner{}
	s := scheduler.New(runner, 50*time.Millisecond, testLogger())
	done := make(chan struct{})

	go s.Start(done)
	time.Sleep(180 * time.Millisecond)
	close(done)
	time.Sleep(20 * time.Millisecond)

	count := atomic.LoadInt64(&runner.count)
	if count < 3 {
		t.Errorf("expected at least 3 runs, got %d", count)
	}
}

func TestScheduler_ContinuesOnRunnerError(t *testing.T) {
	runner := &countingRunner{err: errors.New("transient error")}
	s := scheduler.New(runner, 40*time.Millisecond, testLogger())
	done := make(chan struct{})

	go s.Start(done)
	time.Sleep(130 * time.Millisecond)
	close(done)
	time.Sleep(20 * time.Millisecond)

	count := atomic.LoadInt64(&runner.count)
	if count < 2 {
		t.Errorf("expected scheduler to continue despite errors, got %d runs", count)
	}
}
