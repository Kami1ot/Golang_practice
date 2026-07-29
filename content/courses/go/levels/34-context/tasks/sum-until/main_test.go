package main

import (
	"context"
	"testing"
	"time"
)

func sumWithTimeout(t *testing.T, ctx context.Context, ch <-chan int) int {
	t.Helper()
	res := make(chan int, 1)
	go func() {
		res <- SumUntil(ctx, ch)
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(3 * time.Second):
		t.Fatalf("SumUntil не завершилась — она должна реагировать и на закрытие канала, и на отмену контекста")
		return 0
	}
}

func TestSumUntilClosedChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	got := sumWithTimeout(t, context.Background(), ch)
	if got != 6 {
		t.Errorf("SumUntil до закрытия канала = %d, ожидалось 6", got)
	}
}

func TestSumUntilCancel(t *testing.T) {
	ch := make(chan int) // без буфера: отправка завершается только после приёма
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch <- 1
		ch <- 2
		cancel() // оба значения гарантированно приняты — теперь отмена
	}()
	got := sumWithTimeout(t, ctx, ch)
	if got != 3 {
		t.Errorf("SumUntil с отменой после 1 и 2 = %d, ожидалось 3", got)
	}
}

func TestSumUntilAlreadyCancelled(t *testing.T) {
	ch := make(chan int) // никто никогда не отправит
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := sumWithTimeout(t, ctx, ch)
	if got != 0 {
		t.Errorf("SumUntil с уже отменённым контекстом = %d, ожидалось 0 — и без зависания на пустом канале", got)
	}
}

func TestSumUntilEmptyClosed(t *testing.T) {
	ch := make(chan int)
	close(ch)
	got := sumWithTimeout(t, context.Background(), ch)
	if got != 0 {
		t.Errorf("SumUntil пустого закрытого канала = %d, ожидалось 0 — не забывайте ok-идиому: закрытый канал отдаёт нули бесконечно", got)
	}
}
