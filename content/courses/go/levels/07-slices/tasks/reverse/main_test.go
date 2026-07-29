package main

import (
	"fmt"
	"testing"
)

func TestReverseBasic(t *testing.T) {
	got := Reverse([]int{1, 2, 3})
	want := []int{3, 2, 1}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("Reverse([1 2 3]) = %v, ожидалось %v", got, want)
	}
}

func TestReverseSingle(t *testing.T) {
	got := Reverse([]int{7})
	if fmt.Sprint(got) != "[7]" {
		t.Errorf("Reverse([7]) = %v, ожидалось [7]", got)
	}
}

func TestReverseEmpty(t *testing.T) {
	got := Reverse(nil)
	if len(got) != 0 {
		t.Errorf("Reverse(nil) вернул %v, ожидался пустой слайс", got)
	}
}

func TestReverseNegative(t *testing.T) {
	got := Reverse([]int{-1, 0, 5, -7})
	want := []int{-7, 5, 0, -1}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("Reverse([-1 0 5 -7]) = %v, ожидалось %v", got, want)
	}
}

func TestReverseDoesNotModifyOriginal(t *testing.T) {
	original := []int{1, 2, 3, 4}
	Reverse(original)
	if fmt.Sprint(original) != "[1 2 3 4]" {
		t.Errorf("исходный слайс изменился: стал %v, а должен остаться [1 2 3 4] — функция обязана строить НОВЫЙ слайс", original)
	}
}

func TestReverseEvenLength(t *testing.T) {
	got := Reverse([]int{10, 20, 30, 40, 50, 60})
	want := []int{60, 50, 40, 30, 20, 10}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("Reverse([10 20 30 40 50 60]) = %v, ожидалось %v", got, want)
	}
}
