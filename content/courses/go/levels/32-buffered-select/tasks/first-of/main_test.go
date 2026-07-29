package main

import (
	"testing"
	"time"
)

func callWithTimeout(t *testing.T, a, b <-chan string) string {
	t.Helper()
	res := make(chan string, 1)
	go func() {
		res <- FirstOf(a, b)
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(3 * time.Second):
		t.Fatalf("FirstOf зависла, хотя значение уже лежало в одном из каналов — застряли на молчащем канале?")
		return ""
	}
}

func TestFirstOfA(t *testing.T) {
	a := make(chan string, 1)
	b := make(chan string) // молчит
	a <- "из А"
	if got := callWithTimeout(t, a, b); got != "из А" {
		t.Errorf("FirstOf = %q, ожидалось %q", got, "из А")
	}
}

func TestFirstOfB(t *testing.T) {
	a := make(chan string) // молчит
	b := make(chan string, 1)
	b <- "из Б"
	if got := callWithTimeout(t, a, b); got != "из Б" {
		t.Errorf("FirstOf = %q, ожидалось %q — каналы равноправны, Б тоже может выиграть", got, "из Б")
	}
}

func TestFirstOfWaits(t *testing.T) {
	a := make(chan string)
	b := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond) // оба канала сперва молчат
		a <- "поздний ответ"
	}()
	if got := callWithTimeout(t, a, b); got != "поздний ответ" {
		t.Errorf("FirstOf = %q, ожидалось %q — при пустых каналах нужно ждать, а не возвращаться сразу (default тут лишний)", got, "поздний ответ")
	}
}
