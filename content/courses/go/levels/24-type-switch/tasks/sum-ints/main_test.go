package main

import "testing"

func TestSumIntsMixed(t *testing.T) {
	values := []any{1, "два", 3, 4.5, true, 10}
	if got := SumInts(values); got != 14 {
		t.Errorf("SumInts([1, \"два\", 3, 4.5, true, 10]) = %d, ожидалось 14 — складываем только int", got)
	}
}

func TestSumIntsFloatIsNotInt(t *testing.T) {
	if got := SumInts([]any{2.0, 3}); got != 3 {
		t.Errorf("SumInts([2.0, 3]) = %d, ожидалось 3 — 2.0 это float64, точный тип не совпал", got)
	}
}

func TestSumIntsNoInts(t *testing.T) {
	if got := SumInts([]any{"го", 3.14, false, nil}); got != 0 {
		t.Errorf("SumInts без единого int = %d, ожидалось 0", got)
	}
}

func TestSumIntsOnlyInts(t *testing.T) {
	if got := SumInts([]any{-5, 5, 100}); got != 100 {
		t.Errorf("SumInts([-5, 5, 100]) = %d, ожидалось 100 — отрицательные тоже считаются", got)
	}
}

func TestSumIntsEmpty(t *testing.T) {
	if got := SumInts([]any{}); got != 0 {
		t.Errorf("SumInts пустого слайса = %d, ожидалось 0", got)
	}
	if got := SumInts(nil); got != 0 {
		t.Errorf("SumInts(nil) = %d, ожидалось 0", got)
	}
}

func TestSumIntsNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SumInts запаниковала: %v — используйте ok-форму assertion", r)
		}
	}()
	SumInts([]any{nil, struct{}{}, []int{1}, "x"})
}
