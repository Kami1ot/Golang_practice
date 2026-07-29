package main

import "testing"

func TestSumPositive(t *testing.T) {
	if got := Sum(3, 4); got != 7 {
		t.Errorf("Sum(3, 4) = %d, ожидалось 7", got)
	}
}

func TestSumNegative(t *testing.T) {
	if got := Sum(-2, -3); got != -5 {
		t.Errorf("Sum(-2, -3) = %d, ожидалось -5", got)
	}
}

func TestSumZero(t *testing.T) {
	if got := Sum(0, 0); got != 0 {
		t.Errorf("Sum(0, 0) = %d, ожидалось 0", got)
	}
}

func TestSumMixed(t *testing.T) {
	if got := Sum(10, -4); got != 6 {
		t.Errorf("Sum(10, -4) = %d, ожидалось 6", got)
	}
}

func TestSumLarge(t *testing.T) {
	if got := Sum(1000000, 2000000); got != 3000000 {
		t.Errorf("Sum(1000000, 2000000) = %d, ожидалось 3000000", got)
	}
}
