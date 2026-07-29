package main

import (
	"reflect"
	"testing"
	"time"
)

func drainWithTimeout(t *testing.T, ch <-chan int) []int {
	t.Helper()
	res := make(chan []int, 1)
	go func() {
		res <- Drain(ch)
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("Drain зависла или зациклилась — она не должна ждать новых значений (default) и не должна крутиться на закрытом канале (ok-идиома)")
		return nil
	}
}

func TestDrainBuffered(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	got := drainWithTimeout(t, ch)
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Drain = %v, ожидалось [1 2 3] — все накопленные значения в порядке FIFO", got)
	}
}

func TestDrainEmpty(t *testing.T) {
	ch := make(chan int, 5)
	got := drainWithTimeout(t, ch)
	if len(got) != 0 {
		t.Errorf("Drain пустого канала = %v, ожидался пустой слайс", got)
	}
}

func TestDrainTwice(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 7
	first := drainWithTimeout(t, ch)
	if !reflect.DeepEqual(first, []int{7}) {
		t.Errorf("первый Drain = %v, ожидалось [7]", first)
	}
	second := drainWithTimeout(t, ch)
	if len(second) != 0 {
		t.Errorf("второй Drain = %v, ожидался пустой слайс — всё уже забрали", second)
	}
}

func TestDrainClosed(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 4
	ch <- 5
	close(ch)
	got := drainWithTimeout(t, ch)
	if !reflect.DeepEqual(got, []int{4, 5}) {
		t.Errorf("Drain закрытого канала с остатком = %v, ожидалось [4 5] — и без бесконечных нулей: проверяйте ok", got)
	}
}
