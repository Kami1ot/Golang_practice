package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEventuallyFast(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "успел!"
	got, err := Eventually(ch, 5*time.Second)
	if err != nil {
		t.Fatalf("Eventually с готовым значением вернула ошибку %v, ожидался nil", err)
	}
	if got != "успел!" {
		t.Errorf("Eventually = %q, ожидалось %q", got, "успел!")
	}
}

func TestEventuallyTimeout(t *testing.T) {
	ch := make(chan string) // молчит вечно
	start := time.Now()
	got, err := Eventually(ch, 50*time.Millisecond)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatalf("Eventually пустого канала вернула nil-ошибку, ожидался таймаут")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("ошибка %v, ожидалась context.DeadlineExceeded — возвращайте ctx.Err(), а не свою ошибку", err)
	}
	if got != "" {
		t.Errorf("при таймауте значение = %q, ожидалась пустая строка", got)
	}
	if elapsed > 3*time.Second {
		t.Errorf("Eventually ждала %v при таймауте 50мс — таймаут не сработал", elapsed)
	}
}

func TestEventuallyLateButInTime(t *testing.T) {
	ch := make(chan string)
	go func() {
		time.Sleep(30 * time.Millisecond)
		ch <- "чуть позже"
	}()
	got, err := Eventually(ch, 5*time.Second)
	if err != nil {
		t.Fatalf("Eventually вернула ошибку %v, а значение пришло задолго до таймаута", err)
	}
	if got != "чуть позже" {
		t.Errorf("Eventually = %q, ожидалось %q", got, "чуть позже")
	}
}
