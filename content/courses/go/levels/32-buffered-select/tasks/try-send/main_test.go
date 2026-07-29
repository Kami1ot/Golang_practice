package main

import (
	"testing"
	"time"
)

func TestTrySendBuffered(t *testing.T) {
	ch := make(chan int, 2)
	if !TrySend(ch, 1) {
		t.Errorf("TrySend в пустой буфер на 2 места = false, ожидалось true")
	}
	if !TrySend(ch, 2) {
		t.Errorf("TrySend во второе место буфера = false, ожидалось true")
	}
	if TrySend(ch, 3) {
		t.Errorf("TrySend в ПОЛНЫЙ буфер = true, ожидалось false — и никакой блокировки")
	}
	if v := <-ch; v != 1 {
		t.Errorf("из канала пришло %d, ожидалось 1 — порядок FIFO, лишнее не отправляем", v)
	}
	if !TrySend(ch, 3) {
		t.Errorf("TrySend после освобождения места = false, ожидалось true")
	}
}

func TestTrySendNoBlock(t *testing.T) {
	ch := make(chan int) // без буфера и без приёмника: отправка невозможна
	done := make(chan bool, 1)
	go func() {
		done <- TrySend(ch, 42)
	}()
	select {
	case got := <-done:
		if got {
			t.Errorf("TrySend без приёмника = true, ожидалось false")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("TrySend заблокировалась — используйте select с default, а не обычную отправку")
	}
}

func TestTrySendValues(t *testing.T) {
	ch := make(chan int, 3)
	for _, v := range []int{10, 20, 30} {
		if !TrySend(ch, v) {
			t.Fatalf("TrySend(%d) = false при свободном буфере", v)
		}
	}
	for _, want := range []int{10, 20, 30} {
		if got := <-ch; got != want {
			t.Errorf("из канала пришло %d, ожидалось %d", got, want)
		}
	}
}
