package app

import (
	"context"
	"sync"
	"time"
)

type trayDispatchQueue struct {
	mu   sync.Mutex
	jobs chan func()
}

func (q *trayDispatchQueue) submit(ctx context.Context, run func(), post func() error) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case q.jobs <- run:
		if err := post(); err != nil {
			select {
			case <-q.jobs:
			default:
			}
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func dispatchTrayUpdate(ctx context.Context, dispatch func(func()) error, update func()) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := dispatch(update); err == nil {
			return nil
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// Cancel queued work, but drain work already using native UI resources.
func dispatchTask(ctx context.Context, submit func(func()) error, fn func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var mu sync.Mutex
	started, cancelled := false, false
	done := make(chan struct{})
	if err := submit(func() {
		mu.Lock()
		if cancelled {
			mu.Unlock()
			close(done)
			return
		}
		started = true
		mu.Unlock()
		fn()
		close(done)
	}); err != nil {
		mu.Lock()
		cancelled = true
		inFlight := started
		mu.Unlock()
		if inFlight {
			<-done
		}
		return err
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		mu.Lock()
		if !started {
			cancelled = true
			mu.Unlock()
			return ctx.Err()
		}
		mu.Unlock()
		<-done
		return ctx.Err()
	}
}
