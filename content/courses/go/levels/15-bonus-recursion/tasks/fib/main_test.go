package main

import "testing"

func TestFibZero(t *testing.T) {
	if got := Fib(0); got != 0 {
		t.Errorf("Fib(0) = %d, ожидалось 0 — это первый базовый случай", got)
	}
}

func TestFibOne(t *testing.T) {
	if got := Fib(1); got != 1 {
		t.Errorf("Fib(1) = %d, ожидалось 1 — это второй базовый случай", got)
	}
}

func TestFibTwo(t *testing.T) {
	if got := Fib(2); got != 1 {
		t.Errorf("Fib(2) = %d, ожидалось 1 (0 + 1) — первый шаг после базовых случаев", got)
	}
}

func TestFibTen(t *testing.T) {
	if got := Fib(10); got != 55 {
		t.Errorf("Fib(10) = %d, ожидалось 55", got)
	}
}

func TestFibTwenty(t *testing.T) {
	if got := Fib(20); got != 6765 {
		t.Errorf("Fib(20) = %d, ожидалось 6765", got)
	}
}
