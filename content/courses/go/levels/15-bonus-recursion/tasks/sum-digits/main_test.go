package main

import "testing"

func TestSumDigitsZero(t *testing.T) {
	if got := SumDigits(0); got != 0 {
		t.Errorf("SumDigits(0) = %d, ожидалось 0 — ноль однозначный, его должен обработать базовый случай", got)
	}
}

func TestSumDigitsSingle(t *testing.T) {
	if got := SumDigits(7); got != 7 {
		t.Errorf("SumDigits(7) = %d, ожидалось 7 — сумма цифр однозначного числа равна ему самому", got)
	}
}

func TestSumDigitsLong(t *testing.T) {
	if got := SumDigits(12345); got != 15 {
		t.Errorf("SumDigits(12345) = %d, ожидалось 15 (1+2+3+4+5)", got)
	}
}

func TestSumDigitsInnerZeros(t *testing.T) {
	if got := SumDigits(10203); got != 6 {
		t.Errorf("SumDigits(10203) = %d, ожидалось 6 (1+0+2+0+3) — нули внутри числа не должны ломать подсчёт", got)
	}
}
