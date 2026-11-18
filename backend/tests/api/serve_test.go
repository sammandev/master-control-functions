package api_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/api"
)

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestServeDrainsInFlightRequestOnShutdown(t *testing.T) {
	entered := make(chan struct{})
	app := fiber.New()
	app.Get("/ping", func(c fiber.Ctx) error { return c.SendString("pong") })
	app.Get("/slow", func(c fiber.Ctx) error {
		close(entered)
		time.Sleep(300 * time.Millisecond)
		return c.SendString("done")
	})

	addr := freeAddr(t)
	base := fmt.Sprintf("http://%s", addr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	served := make(chan error, 1)
	go func() { served <- api.Serve(ctx, app, addr, 5*time.Second) }()

	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := client.Get(base + "/ping")
		if err == nil {
			resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became ready: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	type result struct {
		status int
		body   string
		err    error
	}
	slow := make(chan result, 1)
	go func() {
		resp, err := client.Get(base + "/slow")
		if err != nil {
			slow <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		slow <- result{status: resp.StatusCode, body: string(b), err: err}
	}()

	<-entered
	cancel()

	r := <-slow
	if r.err != nil || r.status != http.StatusOK || r.body != "done" {
		t.Fatalf("in-flight request not drained: status=%d body=%q err=%v", r.status, r.body, r.err)
	}

	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("Serve did not return after shutdown")
	}

	if conn, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
		conn.Close()
		t.Fatalf("server still accepting connections after shutdown")
	}
}
