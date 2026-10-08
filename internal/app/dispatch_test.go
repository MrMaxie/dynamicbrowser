package app

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTrayDispatchQueueRecoversFromPostFailure(t *testing.T) {
	queue := trayDispatchQueue{jobs: make(chan func(), 1)}
	failure := errors.New("PostMessageW failed")
	if err := dispatchTask(context.Background(), func(run func()) error {
		return queue.submit(context.Background(), run, func() error { return failure })
	}, func() { t.Error("failed submission ran") }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if len(queue.jobs) != 0 {
		t.Fatal("failed submission left cancelled work in the queue")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ran := false
	if err := dispatchTask(ctx, func(run func()) error {
		return queue.submit(ctx, run, func() error { (<-queue.jobs)(); return nil })
	}, func() { ran = true }); err != nil || !ran {
		t.Fatalf("subsequent submission: ran=%v, err=%v", ran, err)
	}
}

func TestTrayDispatchRollbackDoesNotRemoveConcurrentSubmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	queue := trayDispatchQueue{jobs: make(chan func(), 1)}
	posting, failPost := make(chan struct{}), make(chan struct{})
	firstResult := make(chan error, 1)
	go func() {
		firstResult <- queue.submit(ctx, func() {}, func() error {
			close(posting)
			select {
			case <-failPost:
			case <-ctx.Done():
			}
			return errors.New("PostMessageW failed")
		})
	}()
	<-posting
	(<-queue.jobs)()
	secondStarted, secondResult := make(chan struct{}), make(chan error, 1)
	ran := false
	go func() {
		close(secondStarted)
		secondResult <- queue.submit(ctx, func() { ran = true }, func() error { return nil })
	}()
	<-secondStarted
	close(failPost)
	if err := <-firstResult; err == nil {
		t.Fatal("expected failed first submission")
	}
	if err := <-secondResult; err != nil {
		t.Fatal(err)
	}
	select {
	case job := <-queue.jobs:
		job()
	default:
		t.Fatal("rollback removed the concurrent successful submission")
	}
	if !ran {
		t.Fatal("successful submission did not run")
	}
}

func TestTrayUpdateRetriesPostFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	attempts, updates := 0, 0
	err := dispatchTrayUpdate(ctx, func(update func()) error {
		attempts++
		if attempts < 3 {
			return errors.New("PostMessageW failed")
		}
		update()
		return nil
	}, func() { updates++ })
	if err != nil || attempts != 3 || updates != 1 {
		t.Fatalf("attempts=%d, updates=%d, err=%v", attempts, updates, err)
	}
}

func TestTrayUpdateRetryStopsOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := dispatchTrayUpdate(ctx, func(func()) error {
		attempts++
		cancel()
		return errors.New("PostMessageW failed")
	}, func() { t.Fatal("cancelled update ran") })
	if !errors.Is(err, context.Canceled) || attempts != 1 {
		t.Fatalf("attempts=%d, err=%v", attempts, err)
	}
}

func TestDispatchCancelsQueuedWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var queued func()
	submitted := make(chan struct{})
	result := make(chan error, 1)
	ran := false
	go func() {
		result <- dispatchTask(ctx, func(job func()) error { queued = job; close(submitted); return nil }, func() { ran = true })
	}()
	<-submitted
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	queued()
	if ran {
		t.Fatal("cancelled native work ran")
	}
	if err := dispatchTask(ctx, func(func()) error { t.Fatal("submitted after cancellation"); return nil }, func() {}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestDispatchDrainsStartedWork(t *testing.T) {
	for _, submitFails := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancellation", true: "submit failure"}[submitFails], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started, finish := make(chan struct{}), make(chan struct{})
			defer close(finish)
			result := make(chan error, 1)
			go func() {
				result <- dispatchTask(ctx, func(job func()) error {
					go job()
					<-started
					if submitFails {
						return errors.New("submit failed")
					}
					return nil
				}, func() { close(started); <-finish })
			}()
			<-started
			cancel()
			select {
			case err := <-result:
				t.Fatalf("did not drain started work: %v", err)
			case <-time.After(20 * time.Millisecond):
			}
			finish <- struct{}{}
			if err := <-result; err == nil {
				t.Fatal("expected cancellation or submission error")
			}
		})
	}
}
