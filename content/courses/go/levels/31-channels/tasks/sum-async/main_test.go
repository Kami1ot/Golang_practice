package main

import (
	"testing"
	"time"
)

func recvWithTimeout(t *testing.T, ch <-chan int) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatalf("значение так и не пришло по каналу — горутина должна отправить сумму")
		return 0
	}
}

func TestSumAsyncBasic(t *testing.T) {
	if got := recvWithTimeout(t, SumAsync(2, 3)); got != 5 {
		t.Errorf("<-SumAsync(2, 3) = %d, ожидалось 5", got)
	}
}

func TestSumAsyncNegative(t *testing.T) {
	if got := recvWithTimeout(t, SumAsync(-10, 4)); got != -6 {
		t.Errorf("<-SumAsync(-10, 4) = %d, ожидалось -6", got)
	}
}

func TestSumAsyncReturnsImmediately(t *testing.T) {
	done := make(chan struct{})
	var ch <-chan int
	go func() {
		ch = SumAsync(1, 1) // сам вызов не должен блокировать
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("SumAsync заблокировалась — отправлять в канал должна ГОРУТИНА, а не сама функция")
	}
	if got := recvWithTimeout(t, ch); got != 2 {
		t.Errorf("<-SumAsync(1, 1) = %d, ожидалось 2", got)
	}
}

func TestSumAsyncIndependentCalls(t *testing.T) {
	a := SumAsync(1, 2)
	b := SumAsync(10, 20)
	if got := recvWithTimeout(t, b); got != 30 {
		t.Errorf("второй вызов вернул %d, ожидалось 30 — у каждого вызова свой канал", got)
	}
	if got := recvWithTimeout(t, a); got != 3 {
		t.Errorf("первый вызов вернул %d, ожидалось 3", got)
	}
}
