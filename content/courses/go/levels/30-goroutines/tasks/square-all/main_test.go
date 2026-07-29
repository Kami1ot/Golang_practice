package main

import (
	"reflect"
	"testing"
)

func TestSquareAllBasic(t *testing.T) {
	got := SquareAll([]int{3, 5, 7})
	want := []int{9, 25, 49}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SquareAll([3, 5, 7]) = %v, ожидалось %v — порядок должен соответствовать входу", got, want)
	}
}

func TestSquareAllNegative(t *testing.T) {
	got := SquareAll([]int{-4, 0, 2})
	want := []int{16, 0, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SquareAll([-4, 0, 2]) = %v, ожидалось %v", got, want)
	}
}

func TestSquareAllBig(t *testing.T) {
	nums := make([]int, 200)
	want := make([]int, 200)
	for i := range nums {
		nums[i] = i
		want[i] = i * i
	}
	got := SquareAll(nums)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SquareAll на 200 элементах вернул неверный результат — вероятно, функция не дождалась всех горутин (Wait до return!)")
	}
}

func TestSquareAllEmpty(t *testing.T) {
	got := SquareAll([]int{})
	if len(got) != 0 {
		t.Errorf("SquareAll([]) = %v, ожидался пустой слайс", got)
	}
}

func TestSquareAllSingle(t *testing.T) {
	got := SquareAll([]int{12})
	if !reflect.DeepEqual(got, []int{144}) {
		t.Errorf("SquareAll([12]) = %v, ожидалось [144]", got)
	}
}
