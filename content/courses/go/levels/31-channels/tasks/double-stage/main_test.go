package main

import (
	"reflect"
	"testing"
	"time"
)

func feed(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, v := range vals {
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func gather(t *testing.T, ch <-chan int, expectMax int) []int {
	t.Helper()
	var got []int
	timeout := time.After(3 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, v)
			if len(got) > expectMax {
				t.Fatalf("выход отдал больше %d значений — закрывайте out после исчерпания in", expectMax)
			}
		case <-timeout:
			t.Fatalf("выходной канал не закрылся: собрано %v — после range по in нужен close(out)", got)
		}
	}
}

func TestDoubleAllBasic(t *testing.T) {
	got := gather(t, DoubleAll(feed(1, 4, 9)), 3)
	want := []int{2, 8, 18}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DoubleAll(1, 4, 9) отдал %v, ожидалось %v — порядок FIFO сохраняется", got, want)
	}
}

func TestDoubleAllEmpty(t *testing.T) {
	got := gather(t, DoubleAll(feed()), 0)
	if len(got) != 0 {
		t.Errorf("DoubleAll пустого входа отдал %v, ожидался пустой закрытый выход", got)
	}
}

func TestDoubleAllNegative(t *testing.T) {
	got := gather(t, DoubleAll(feed(-3, 0, 5)), 3)
	want := []int{-6, 0, 10}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DoubleAll(-3, 0, 5) отдал %v, ожидалось %v", got, want)
	}
}

func TestDoubleAllChained(t *testing.T) {
	got := gather(t, DoubleAll(DoubleAll(feed(1, 2, 3))), 3)
	want := []int{4, 8, 12}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("цепочка DoubleAll(DoubleAll(...)) отдала %v, ожидалось %v — стадии должны сцепляться", got, want)
	}
}
