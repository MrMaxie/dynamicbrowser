package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MrMaxie/dynamicbrowser/internal/config"
	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

func TestRoutingControl(t *testing.T) {
	requests := make(chan routing.Request, 2)
	stop, err := startRoutingControl(func(request routing.Request) error {
		requests <- request
		if request.URLs[0] == "https://example.com/error" {
			return errors.New("configure a valid default")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	request := routing.Request{URLs: []string{"https://example.com/a", "https://example.com/b"}, Source: routing.Source{Process: "Editor", Window: "Document"}}
	if err := forwardURLs(request); err != nil {
		t.Fatal(err)
	}
	if got := <-requests; !reflect.DeepEqual(got, request) {
		t.Fatalf("request = %#v", got)
	}
	if err := forwardURLs(routing.Request{URLs: []string{"https://example.com/error"}}); err == nil || err.Error() != "configure a valid default" {
		t.Fatalf("response = %v", err)
	}
	<-requests
	stop()
	if err := forwardURLs(request); !errors.Is(err, errRestartNotReady) {
		t.Fatalf("stopped listener: %v", err)
	}
}

func TestRoutingControlBatches(t *testing.T) {
	for _, tt := range []struct {
		name string
		urls []string
	}{
		{"count", make([]string, 513)},
		{"encoded size", []string{"https://example.com/" + strings.Repeat(`\"`, maxRoutingBytes/6), "https://example.com/" + strings.Repeat(`\"`, maxRoutingBytes/6)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "count" {
				for i := range tt.urls {
					tt.urls[i] = fmt.Sprintf("https://example.com/%d", i)
				}
			}
			source := routing.Source{Process: "Editor", Window: "Document"}
			var received []string
			calls := 0
			stop, err := startRoutingControl(func(request routing.Request) error {
				if request.Source != source {
					t.Error("source changed between batches")
				}
				received = append(received, request.URLs...)
				calls++
				if calls == 1 {
					return errors.New("first batch contains an unavailable browser")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			err = forwardURLs(routing.Request{URLs: tt.urls, Source: source})
			stop()
			if err == nil || !strings.Contains(err.Error(), "unavailable browser") {
				t.Fatalf("routing error = %v", err)
			}
			if calls < 2 || !reflect.DeepEqual(received, tt.urls) {
				t.Fatalf("received %d of %d URLs in %d batches", len(received), len(tt.urls), calls)
			}
		})
	}
}

func TestRoutingControlOversizedURLDoesNotPartiallySend(t *testing.T) {
	calls := 0
	stop, err := startRoutingControl(func(routing.Request) error { calls++; return nil })
	if err != nil {
		t.Fatal(err)
	}
	err = forwardURLs(routing.Request{URLs: []string{"https://example.com/ok", "https://example.com/" + strings.Repeat("a", maxRoutingBytes)}})
	stop()
	if err == nil || calls != 0 {
		t.Fatalf("oversized request: calls=%d, err=%v", calls, err)
	}
}

func TestRoutingControlDoesNotRetryAcknowledgedBatches(t *testing.T) {
	listener, err := listenRouting()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	done := make(chan struct{})
	var received []string
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		handleRouting(conn, func(request routing.Request) error {
			received = append(received, request.URLs...)
			return listener.Close()
		})
	}()
	urls := make([]string, maxRoutingURLs+1)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://example.com/%d", i)
	}
	err = forwardURLs(routing.Request{URLs: urls})
	<-done
	if err == nil || errors.Is(err, errRestartNotReady) {
		t.Fatalf("partial handoff may be retried by startup: %v", err)
	}
	if !reflect.DeepEqual(received, urls[:maxRoutingURLs]) {
		t.Fatalf("unexpected acknowledged URLs: %d", len(received))
	}
}

func TestRoutingBatchByteLimit(t *testing.T) {
	request := routing.Request{URLs: []string{"https://example.com/" + strings.Repeat(`\"`, maxRoutingBytes/6), "https://example.com/" + strings.Repeat(`\"`, maxRoutingBytes/6)}, Source: routing.Source{Window: `quoted "window"`}}
	batches, err := routingBatches(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 2 {
		t.Fatalf("batch count = %d", len(batches))
	}
	for _, batch := range batches {
		encoded, err := json.Marshal(batch)
		if err != nil || len(encoded)+1 > maxRoutingBytes || batch.Source != request.Source {
			t.Fatalf("invalid batch: bytes=%d, err=%v", len(encoded)+1, err)
		}
	}
}

func TestRoutingBatchExactByteBoundary(t *testing.T) {
	request := routing.Request{URLs: []string{"https://example.com/"}, Source: routing.Source{Window: `quoted "window"`}}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	request.URLs[0] += strings.Repeat("a", maxRoutingBytes-len(encoded)-1)
	request.URLs = append(request.URLs, "https://example.com/next")
	batches, err := routingBatches(request)
	if err != nil || len(batches) != 2 {
		t.Fatalf("exact boundary: batches=%d, err=%v", len(batches), err)
	}
	encoded, err = json.Marshal(batches[0])
	if err != nil || len(encoded)+1 != maxRoutingBytes {
		t.Fatalf("boundary bytes=%d, err=%v", len(encoded)+1, err)
	}
	request.URLs[0] += "a"
	if batches, err := routingBatches(request); err == nil || len(batches) != 0 {
		t.Fatal("oversized URL was not rejected before batching")
	}
}

func TestRoutingRejectsInvalidRequests(t *testing.T) {
	for _, payload := range []string{`{}`, `{"URLs":[],"Unknown":true}`, `not json`} {
		server, client := net.Pipe()
		done := make(chan struct{})
		go func() {
			handleRouting(server, func(routing.Request) error { t.Error("invalid request was routed"); return nil })
			close(done)
		}()
		_ = client.SetDeadline(time.Now().Add(time.Second))
		if _, err := client.Write([]byte(payload + "\n")); err != nil {
			t.Fatal(err)
		}
		var response string
		if err := json.NewDecoder(client).Decode(&response); err != nil || response == "" {
			t.Fatalf("invalid request: %q, %v", response, err)
		}
		_ = client.Close()
		<-done
	}
}

func TestTraySelection(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Browsers: config.Browsers{{Name: "first", Exe: executable}, {Name: "second", Exe: executable}, {Name: "missing", Exe: "nonexistent-browser-test"}}}
	router := routing.New(cfg, t.TempDir(), func(string, []string) error { return nil })
	view := newTrayMenu(context.Background(), router, func() error { return nil }, func() {})
	if len(view.checks) != 3 || !view.checks[""].IsChecked() {
		t.Fatal("menu does not start in Auto with available browsers")
	}
	for _, choice := range []string{"first", "second", ""} {
		view.choose(choice)
		if router.Forced() != choice {
			t.Fatal("selection did not change routing")
		}
		for name, item := range view.checks {
			if item.IsChecked() != (name == choice) {
				t.Fatalf("check for %q after selecting %q", name, choice)
			}
		}
	}
	view.choose("second")
	router.Update(&config.Config{Browsers: cfg.Browsers[:1]})
	view = newTrayMenu(context.Background(), router, func() error { return nil }, func() {})
	if router.Forced() != "" || !view.checks[""].IsChecked() {
		t.Fatal("removed browser did not reset Auto")
	}
	ctx, cancel := context.WithCancel(context.Background())
	stale := newTrayMenu(ctx, router, func() error { return nil }, func() {})
	cancel()
	stale.choose("first")
	if router.Forced() != "" {
		t.Fatal("retired menu changed routing")
	}
}
