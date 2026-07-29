package main

import "testing"

func TestStatsEmpty(t *testing.T) {
	min, max, sum := Stats()
	if min != 0 || max != 0 || sum != 0 {
		t.Errorf("Stats() = (%d, %d, %d), ожидалось (0, 0, 0) — вызов без аргументов должен вернуть нули", min, max, sum)
	}
}

func TestStatsSingle(t *testing.T) {
	min, max, sum := Stats(7)
	if min != 7 || max != 7 || sum != 7 {
		t.Errorf("Stats(7) = (%d, %d, %d), ожидалось (7, 7, 7) — единственное число сразу и минимум, и максимум, и сумма", min, max, sum)
	}
}

func TestStatsRegular(t *testing.T) {
	min, max, sum := Stats(3, 1, 4, 1, 5)
	if min != 1 || max != 5 || sum != 14 {
		t.Errorf("Stats(3, 1, 4, 1, 5) = (%d, %d, %d), ожидалось (1, 5, 14)", min, max, sum)
	}
}

func TestStatsNegative(t *testing.T) {
	min, max, sum := Stats(-2, -8, -5)
	if min != -8 || max != -2 || sum != -15 {
		t.Errorf("Stats(-2, -8, -5) = (%d, %d, %d), ожидалось (-8, -2, -15) — проверьте старт поиска: инициализация min и max нулём подводит на отрицательных числах", min, max, sum)
	}
}

func TestStatsUnpackSlice(t *testing.T) {
	data := []int{10, -10, 20}
	min, max, sum := Stats(data...)
	if min != -10 || max != 20 || sum != 20 {
		t.Errorf("Stats(data...) для data = []int{10, -10, 20} = (%d, %d, %d), ожидалось (-10, 20, 20) — распакованный слайс должен работать как обычные аргументы", min, max, sum)
	}
}
