package main

import "testing"

func TestStackLIFO(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Push(2)
	s.Push(3)
	want := []int{3, 2, 1}
	for _, w := range want {
		got, ok := s.Pop()
		if !ok {
			t.Fatalf("Pop() вернул ok=false, хотя в стеке ещё оставались элементы — проверьте, что Push с получателем-указателем реально добавляет в s.items")
		}
		if got != w {
			t.Errorf("Pop() = %d, ожидалось %d — стек отдаёт элементы в обратном порядке (LIFO)", got, w)
		}
	}
}

func TestStackPopEmpty(t *testing.T) {
	var s Stack
	got, ok := s.Pop()
	if ok {
		t.Errorf("Pop() пустого стека вернул ok=true, ожидалось false")
	}
	if got != 0 {
		t.Errorf("Pop() пустого стека вернул %d, ожидалось 0", got)
	}
}

func TestStackLen(t *testing.T) {
	var s Stack
	if got := s.Len(); got != 0 {
		t.Errorf("Len() нового стека = %d, ожидалось 0", got)
	}
	s.Push(10)
	s.Push(20)
	s.Push(30)
	if got := s.Len(); got != 3 {
		t.Errorf("Len() после трёх Push = %d, ожидалось 3", got)
	}
	s.Pop()
	if got := s.Len(); got != 2 {
		t.Errorf("Len() после Pop = %d, ожидалось 2 — Pop должен удалять элемент из стека, а не только читать", got)
	}
}

func TestStackPushAfterPop(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Pop()
	s.Push(42)
	got, ok := s.Pop()
	if !ok || got != 42 {
		t.Errorf("Push(42) после Pop: Pop() = (%d, %v), ожидалось (42, true) — стек должен работать и после снятий", got, ok)
	}
	if s.Len() != 0 {
		t.Errorf("Len() после равного числа Push и Pop = %d, ожидалось 0", s.Len())
	}
}
