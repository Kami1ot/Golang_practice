package main

import (
	"math"
	"testing"
)

func TestTotalMixed(t *testing.T) {
	values := []any{1, 2.5, "три", true}
	if got := Total(values); math.Abs(got-3.5) > 1e-9 {
		t.Errorf("Total([1, 2.5, \"три\", true]) = %v, ожидалось 3.5", got)
	}
}

func TestTotalIntsOnly(t *testing.T) {
	if got := Total([]any{1, 2, 3}); math.Abs(got-6) > 1e-9 {
		t.Errorf("Total([1, 2, 3]) = %v, ожидалось 6 — целые конвертируются в float64", got)
	}
}

func TestTotalFloatsOnly(t *testing.T) {
	if got := Total([]any{0.5, 1.25}); math.Abs(got-1.75) > 1e-9 {
		t.Errorf("Total([0.5, 1.25]) = %v, ожидалось 1.75", got)
	}
}

func TestTotalNothingToSum(t *testing.T) {
	if got := Total([]any{"го", false, nil}); got != 0 {
		t.Errorf("Total без чисел = %v, ожидалось 0", got)
	}
	if got := Total(nil); got != 0 {
		t.Errorf("Total(nil) = %v, ожидалось 0", got)
	}
}

func TestTotalNegative(t *testing.T) {
	if got := Total([]any{-2, 5.5, -0.5}); math.Abs(got-3) > 1e-9 {
		t.Errorf("Total([-2, 5.5, -0.5]) = %v, ожидалось 3", got)
	}
}

func TestTotalNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Total запаниковала: %v — разберите типы безопасно", r)
		}
	}()
	Total([]any{nil, struct{}{}, []float64{1.5}, "x"})
}
