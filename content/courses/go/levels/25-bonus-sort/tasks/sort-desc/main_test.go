package main

import (
	"reflect"
	"testing"
)

func TestSortDescBasic(t *testing.T) {
	nums := []int{40, 95, 60, 95}
	SortDesc(nums)
	want := []int{95, 95, 60, 40}
	if !reflect.DeepEqual(nums, want) {
		t.Errorf("SortDesc([40, 95, 60, 95]) дал %v, ожидалось %v", nums, want)
	}
}

func TestSortDescNegative(t *testing.T) {
	nums := []int{-1, 5, 0, -10}
	SortDesc(nums)
	want := []int{5, 0, -1, -10}
	if !reflect.DeepEqual(nums, want) {
		t.Errorf("SortDesc([-1, 5, 0, -10]) дал %v, ожидалось %v", nums, want)
	}
}

func TestSortDescAlreadySorted(t *testing.T) {
	nums := []int{9, 7, 3}
	SortDesc(nums)
	want := []int{9, 7, 3}
	if !reflect.DeepEqual(nums, want) {
		t.Errorf("SortDesc уже убывающего слайса дал %v, ожидалось %v", nums, want)
	}
}

func TestSortDescEdge(t *testing.T) {
	empty := []int{}
	SortDesc(empty)
	if len(empty) != 0 {
		t.Errorf("SortDesc пустого слайса что-то в него добавил: %v", empty)
	}
	single := []int{42}
	SortDesc(single)
	if !reflect.DeepEqual(single, []int{42}) {
		t.Errorf("SortDesc([42]) дал %v, ожидалось [42]", single)
	}
}
