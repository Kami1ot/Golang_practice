package main

import (
	"context"
	"testing"
	"time"
)

func statusWithTimeout(t *testing.T, ctx context.Context) string {
	t.Helper()
	res := make(chan string, 1)
	go func() {
		res <- Status(ctx)
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("Status заблокировалась — для живого контекста нужен default, а не ожидание Done")
		return ""
	}
}

func TestStatusAlive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if got := statusWithTimeout(t, ctx); got != "работаем" {
		t.Errorf("Status живого контекста = %q, ожидалось %q", got, "работаем")
	}
}

func TestStatusCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := statusWithTimeout(t, ctx); got != "отменено" {
		t.Errorf("Status отменённого контекста = %q, ожидалось %q", got, "отменено")
	}
}

func TestStatusBackground(t *testing.T) {
	if got := statusWithTimeout(t, context.Background()); got != "работаем" {
		t.Errorf("Status(Background) = %q, ожидалось %q — корневой контекст не отменяется никогда", got, "работаем")
	}
}

func TestStatusTransition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if got := statusWithTimeout(t, ctx); got != "работаем" {
		t.Fatalf("до отмены: %q, ожидалось %q", got, "работаем")
	}
	cancel()
	if got := statusWithTimeout(t, ctx); got != "отменено" {
		t.Errorf("после отмены: %q, ожидалось %q", got, "отменено")
	}
}
