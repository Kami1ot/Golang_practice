package main

import "testing"

func TestDivideOK(t *testing.T) {
	got, err := Divide(10, 2)
	if err != nil {
		t.Fatalf("Divide(10, 2) вернула ошибку %v, ожидался nil", err)
	}
	if got != 5 {
		t.Errorf("Divide(10, 2) = %d, ожидалось 5", got)
	}
}

func TestDivideTruncates(t *testing.T) {
	got, err := Divide(7, 2)
	if err != nil {
		t.Fatalf("Divide(7, 2) вернула ошибку %v, ожидался nil", err)
	}
	if got != 3 {
		t.Errorf("Divide(7, 2) = %d, ожидалось 3 — деление целочисленное", got)
	}
}

func TestDivideNegative(t *testing.T) {
	got, err := Divide(-9, 3)
	if err != nil {
		t.Fatalf("Divide(-9, 3) вернула ошибку %v, ожидался nil", err)
	}
	if got != -3 {
		t.Errorf("Divide(-9, 3) = %d, ожидалось -3", got)
	}
}

func TestDivideByZero(t *testing.T) {
	got, err := Divide(1, 0)
	if err == nil {
		t.Fatalf("Divide(1, 0) вернула nil, ожидалась ошибка")
	}
	if err.Error() != "деление на ноль" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "деление на ноль")
	}
	if got != 0 {
		t.Errorf("при ошибке первый результат = %d, ожидалось нулевое значение 0", got)
	}
}
