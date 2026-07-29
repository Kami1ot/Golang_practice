package main

import "testing"

func TestDoublePositive(t *testing.T) {
	n := 21
	Double(&n)
	if n != 42 {
		t.Errorf("после Double(&n) при n = 21 в переменной оказалось %d, ожидалось 42. Функция должна менять значение по указателю", n)
	}
}

func TestDoubleZero(t *testing.T) {
	n := 0
	Double(&n)
	if n != 0 {
		t.Errorf("после Double(&n) при n = 0 в переменной оказалось %d, ожидалось 0", n)
	}
}

func TestDoubleNegative(t *testing.T) {
	n := -7
	Double(&n)
	if n != -14 {
		t.Errorf("после Double(&n) при n = -7 в переменной оказалось %d, ожидалось -14 — отрицательные тоже удваиваются", n)
	}
}

func TestDoubleTwice(t *testing.T) {
	n := 3
	Double(&n)
	Double(&n)
	if n != 12 {
		t.Errorf("после двух вызовов Double(&n) при n = 3 в переменной оказалось %d, ожидалось 12 (3 → 6 → 12)", n)
	}
}
