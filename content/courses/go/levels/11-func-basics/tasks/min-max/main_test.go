package main

import "testing"

func TestMinMaxOrdered(t *testing.T) {
	lo, hi := MinMax(1, 2, 3)
	if lo != 1 || hi != 3 {
		t.Errorf("MinMax(1, 2, 3) = (%d, %d), ожидалось (1, 3) — сначала минимум, потом максимум", lo, hi)
	}
}

func TestMinMaxPermutations(t *testing.T) {
	perms := [][]int{
		{1, 2, 3}, {1, 3, 2}, {2, 1, 3}, {2, 3, 1}, {3, 1, 2}, {3, 2, 1},
	}
	for _, p := range perms {
		lo, hi := MinMax(p[0], p[1], p[2])
		if lo != 1 || hi != 3 {
			t.Errorf("MinMax(%d, %d, %d) = (%d, %d), ожидалось (1, 3) — результат не должен зависеть от порядка аргументов", p[0], p[1], p[2], lo, hi)
		}
	}
}

func TestMinMaxAllEqual(t *testing.T) {
	lo, hi := MinMax(7, 7, 7)
	if lo != 7 || hi != 7 {
		t.Errorf("MinMax(7, 7, 7) = (%d, %d), ожидалось (7, 7) — когда все числа равны, минимум и максимум совпадают", lo, hi)
	}
}

func TestMinMaxNegative(t *testing.T) {
	lo, hi := MinMax(-5, -1, -3)
	if lo != -5 || hi != -1 {
		t.Errorf("MinMax(-5, -1, -3) = (%d, %d), ожидалось (-5, -1) — проверьте работу с отрицательными числами", lo, hi)
	}
}

func TestMinMaxMixedSigns(t *testing.T) {
	lo, hi := MinMax(10, -10, 0)
	if lo != -10 || hi != 10 {
		t.Errorf("MinMax(10, -10, 0) = (%d, %d), ожидалось (-10, 10)", lo, hi)
	}
}

func TestMinMaxTwoEqualMin(t *testing.T) {
	lo, hi := MinMax(2, 2, 5)
	if lo != 2 || hi != 5 {
		t.Errorf("MinMax(2, 2, 5) = (%d, %d), ожидалось (2, 5) — два равных минимума не должны ломать ответ", lo, hi)
	}
	lo, hi = MinMax(4, 9, 9)
	if lo != 4 || hi != 9 {
		t.Errorf("MinMax(4, 9, 9) = (%d, %d), ожидалось (4, 9) — два равных максимума не должны ломать ответ", lo, hi)
	}
}
