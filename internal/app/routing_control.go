package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/MrMaxie/dynamicbrowser/internal/routing"
)

const (
	maxRoutingBytes = 1 << 20
	maxRoutingURLs  = 256
)

func startRoutingControl(open func(routing.Request) error) (func(), error) {
	listener, err := listenRouting()
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			handleRouting(conn, open)
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() { _ = listener.Close(); <-done })
	}, nil
}

func handleRouting(conn net.Conn, open func(routing.Request) error) {
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	var request routing.Request
	decoder := json.NewDecoder(io.LimitReader(conn, maxRoutingBytes))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&request)
	if err == nil {
		if len(request.URLs) == 0 || len(request.URLs) > maxRoutingURLs {
			err = fmt.Errorf("a routing request must contain between 1 and %d URLs", maxRoutingURLs)
		} else {
			err = open(request)
		}
	}
	response := ""
	if err != nil {
		response = err.Error()
	}
	_ = json.NewEncoder(conn).Encode(response)
}

func forwardURLs(request routing.Request) error {
	batches, err := routingBatches(request)
	if err != nil {
		return err
	}
	var failures []error
	for i, batch := range batches {
		acknowledged, err := forwardRoutingBatch(batch)
		if err != nil {
			if !acknowledged {
				if i == 0 {
					return err
				}
				// Startup must not retry already acknowledged URLs after losing the tray.
				return errors.Join(append(failures, fmt.Errorf("tray handoff interrupted after %d batches: %v", i, err))...)
			}
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func routingBatches(request routing.Request) ([]routing.Request, error) {
	if len(request.URLs) == 0 {
		return nil, errors.New("a routing request must contain at least one URL")
	}
	empty, _ := json.Marshal(routing.Request{URLs: []string{}, Source: request.Source})
	overhead := len(empty) + 1
	var batches []routing.Request
	start, size := 0, overhead
	for i, url := range request.URLs {
		encoded, _ := json.Marshal(url)
		if overhead+len(encoded) > maxRoutingBytes {
			return nil, fmt.Errorf("URL %d and its source exceed the routing request size limit", i+1)
		}
		added := len(encoded)
		if i > start {
			added++
		}
		if i-start == maxRoutingURLs || size+added > maxRoutingBytes {
			batches = append(batches, routing.Request{URLs: request.URLs[start:i], Source: request.Source})
			start, size, added = i, overhead, len(encoded)
		}
		size += added
	}
	return append(batches, routing.Request{URLs: request.URLs[start:], Source: request.Source}), nil
}

func forwardRoutingBatch(request routing.Request) (bool, error) {
	conn, err := dialRouting()
	if err != nil {
		return false, err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return false, err
	}
	var response string
	if err := json.NewDecoder(io.LimitReader(conn, maxRoutingBytes)).Decode(&response); err != nil {
		return false, err
	}
	if response != "" {
		return true, errors.New(response)
	}
	return true, nil
}
