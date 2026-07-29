package main

import (
	"testing"
	"time"
)

func filled(vals ...int) chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func sumPoolWithTimeout(t *testing.T, jobs <-chan int, workers int) int {
	t.Helper()
	res := make(chan int, 1)
	go func() {
		res <- SumPool(jobs, workers)
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(3 * time.Second):
		t.Fatalf("SumPool зависла — дождались ли вы всех воркеров и не закрыли ли что-нибудь лишний раз?")
		return 0
	}
}

func TestSumPoolBasic(t *testing.T) {
	if got := sumPoolWithTimeout(t, filled(1, 2, 3, 4, 5), 3); got != 15 {
		t.Errorf("SumPool(1..5, 3 воркера) = %d, ожидалось 15", got)
	}
}

func TestSumPoolOneWorker(t *testing.T) {
	if got := sumPoolWithTimeout(t, filled(10, 20, 30), 1); got != 60 {
		t.Errorf("SumPool с одним воркером = %d, ожидалось 60", got)
	}
}

func TestSumPoolMoreWorkersThanJobs(t *testing.T) {
	if got := sumPoolWithTimeout(t, filled(7), 8); got != 7 {
		t.Errorf("SumPool([7], 8 воркеров) = %d, ожидалось 7", got)
	}
}

func TestSumPoolEmpty(t *testing.T) {
	if got := sumPoolWithTimeout(t, filled(), 4); got != 0 {
		t.Errorf("SumPool пустого канала = %d, ожидалось 0", got)
	}
}

func TestSumPoolBigNoRace(t *testing.T) {
	const n = 5000
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- 1
	}
	close(ch)
	if got := sumPoolWithTimeout(t, ch, 8); got != n {
		t.Errorf("SumPool на %d единицах = %d, ожидалось %d — слагаемые теряются: защитите общий счётчик", n, got, n)
	}
}
