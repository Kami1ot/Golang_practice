package main

import "testing"

func TestHalfSumsEven(t *testing.T) {
	l, r := HalfSums([]int{1, 2, 3, 4})
	if l != 3 || r != 7 {
		t.Errorf("HalfSums([1, 2, 3, 4]) = (%d, %d), ожидалось (3, 7)", l, r)
	}
}

func TestHalfSumsOdd(t *testing.T) {
	l, r := HalfSums([]int{1, 2, 3, 4, 5})
	if l != 3 || r != 12 {
		t.Errorf("HalfSums([1, 2, 3, 4, 5]) = (%d, %d), ожидалось (3, 12) — при нечётной длине лишний элемент справа", l, r)
	}
}

func TestHalfSumsEmpty(t *testing.T) {
	l, r := HalfSums([]int{})
	if l != 0 || r != 0 {
		t.Errorf("HalfSums([]) = (%d, %d), ожидалось (0, 0)", l, r)
	}
}

func TestHalfSumsSingle(t *testing.T) {
	l, r := HalfSums([]int{10})
	if l != 0 || r != 10 {
		t.Errorf("HalfSums([10]) = (%d, %d), ожидалось (0, 10) — единственный элемент уходит вправо", l, r)
	}
}

func TestHalfSumsNegative(t *testing.T) {
	l, r := HalfSums([]int{-5, 5, -3, 3})
	if l != 0 || r != 0 {
		t.Errorf("HalfSums([-5, 5, -3, 3]) = (%d, %d), ожидалось (0, 0)", l, r)
	}
}

func TestHalfSumsBig(t *testing.T) {
	nums := make([]int, 2000)
	for i := range nums {
		nums[i] = 1
	}
	l, r := HalfSums(nums)
	if l != 1000 || r != 1000 {
		t.Errorf("HalfSums на 2000 единиц = (%d, %d), ожидалось (1000, 1000) — похоже, функция вернулась до Wait", l, r)
	}
}
