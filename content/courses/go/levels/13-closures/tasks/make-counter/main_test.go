package main

import "testing"

func TestCounterSequence(t *testing.T) {
	next := MakeCounter()
	if next == nil {
		t.Fatal("MakeCounter() вернул nil, ожидалась функция-счётчик")
	}
	for want := 1; want <= 5; want++ {
		got := next()
		if got != want {
			t.Fatalf("вызов №%d счётчика вернул %d, ожидалось %d", want, got, want)
		}
	}
}

func TestCountersIndependent(t *testing.T) {
	a := MakeCounter()
	b := MakeCounter()
	if a == nil || b == nil {
		t.Fatal("MakeCounter() вернул nil, ожидалась функция-счётчик")
	}
	a()
	a()
	if got := a(); got != 3 {
		t.Errorf("третий вызов первого счётчика вернул %d, ожидалось 3", got)
	}
	if got := b(); got != 1 {
		t.Errorf("первый вызов второго счётчика вернул %d, ожидалось 1 — у каждого вызова MakeCounter должно быть СВОЁ окружение", got)
	}
	if got := a(); got != 4 {
		t.Errorf("четвёртый вызов первого счётчика вернул %d, ожидалось 4 — вызовы второго счётчика не должны влиять на первый", got)
	}
}
