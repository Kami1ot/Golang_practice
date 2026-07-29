package main

import "testing"

func TestApplyDouble(t *testing.T) {
	got := Apply([]int{1, 2, 3}, func(x int) int { return x * 2 })
	want := []int{2, 4, 6}
	if len(got) != len(want) {
		t.Fatalf("Apply([1 2 3], удвоение) вернул слайс длины %d, ожидалась длина %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Apply([1 2 3], удвоение): элемент №%d равен %d, ожидалось %d", i, got[i], want[i])
		}
	}
}

func TestApplySquare(t *testing.T) {
	got := Apply([]int{-2, 0, 5}, func(x int) int { return x * x })
	want := []int{4, 0, 25}
	if len(got) != len(want) {
		t.Fatalf("Apply([-2 0 5], квадрат) вернул слайс длины %d, ожидалась длина %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Apply([-2 0 5], квадрат): элемент №%d равен %d, ожидалось %d", i, got[i], want[i])
		}
	}
}

func TestApplyEmpty(t *testing.T) {
	got := Apply([]int{}, func(x int) int { return x + 1 })
	if len(got) != 0 {
		t.Errorf("Apply на пустом слайсе вернул %v, ожидался пустой результат", got)
	}
}

func TestApplyKeepsSource(t *testing.T) {
	nums := []int{1, 2, 3}
	Apply(nums, func(x int) int { return x * 100 })
	want := []int{1, 2, 3}
	for i := range want {
		if nums[i] != want[i] {
			t.Errorf("после Apply исходный слайс изменился: nums = %v, ожидалось [1 2 3]. Стройте НОВЫЙ слайс, не пишите в nums", nums)
			return
		}
	}
}
