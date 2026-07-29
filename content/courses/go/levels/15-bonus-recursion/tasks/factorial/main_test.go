package main

import "testing"

func TestFactorialZero(t *testing.T) {
	if got := Factorial(0); got != 1 {
		t.Errorf("Factorial(0) = %d, ожидалось 1 — это базовый случай: 0! равен 1 по определению", got)
	}
}

func TestFactorialOne(t *testing.T) {
	if got := Factorial(1); got != 1 {
		t.Errorf("Factorial(1) = %d, ожидалось 1", got)
	}
}

func TestFactorialFive(t *testing.T) {
	if got := Factorial(5); got != 120 {
		t.Errorf("Factorial(5) = %d, ожидалось 120 (5·4·3·2·1)", got)
	}
}

func TestFactorialTwelve(t *testing.T) {
	if got := Factorial(12); got != 479001600 {
		t.Errorf("Factorial(12) = %d, ожидалось 479001600 — проверьте, что рекурсия доходит до конца", got)
	}
}
