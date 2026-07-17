package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServeHTTPServerServesRequestsAndGracefullyStops(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	stopping := make(chan struct{})
	go func() {
		result <- serveHTTPServer(ctx, server, listener, time.Second, func() { close(stopping) })
	}()
	response, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || string(body) != "ok" {
		t.Fatalf("response = %q, err = %v", body, err)
	}
	cancel()
	select {
	case <-stopping:
	case <-time.After(time.Second):
		t.Fatal("stopping callback did not run")
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop")
	}
}

func TestServeHTTPServerReturnsServeErrors(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = serveHTTPServer(ctx, server, listener, time.Second, nil)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("err = %v, want serve error", err)
	}
}

func TestServeHTTPServerShutdownBound(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { once.Do(func() { close(started) }); <-release })}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- serveHTTPServer(ctx, server, listener, 20*time.Millisecond, nil) }()
	requestDone := make(chan error, 1)
	go func() { _, err := http.Get("http://" + listener.Addr().String()); requestDone <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("expected shutdown timeout")
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown exceeded bound")
	}
	close(release)
	<-requestDone
}
