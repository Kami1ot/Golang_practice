package main

import "testing"

func TestCountBasic(t *testing.T) {
	got := Count([]int{3, 1, 3})
	if got[3] != 2 {
		t.Errorf("Count([3 1 3])[3] = %d, ожидалось 2", got[3])
	}
	if got[1] != 1 {
		t.Errorf("Count([3 1 3])[1] = %d, ожидалось 1", got[1])
	}
}

func TestCountLen(t *testing.T) {
	got := Count([]int{5, 5, 5, 2, 2, 9})
	if len(got) != 3 {
		t.Errorf("в Count([5 5 5 2 2 9]) должно быть 3 ключа (5, 2, 9), а карта содержит %d: %v", len(got), got)
	}
}

func TestCountMissingKey(t *testing.T) {
	got := Count([]int{1, 2, 3})
	if _, ok := got[42]; ok {
		t.Errorf("в карте оказался ключ 42, которого нет в исходном слайсе: %v", got)
	}
}

func TestCountEmptyNotNil(t *testing.T) {
	got := Count(nil)
	if got == nil {
		t.Errorf("Count(nil) вернул nil-карту, ожидалась пустая карта, созданная через make")
	}
	if len(got) != 0 {
		t.Errorf("Count(nil) вернул непустую карту: %v", got)
	}
}

func TestCountNegative(t *testing.T) {
	got := Count([]int{-7, 0, -7, -7, 0})
	if got[-7] != 3 {
		t.Errorf("Count([-7 0 -7 -7 0])[-7] = %d, ожидалось 3", got[-7])
	}
	if got[0] != 2 {
		t.Errorf("Count([-7 0 -7 -7 0])[0] = %d, ожидалось 2", got[0])
	}
}

func TestCountSingle(t *testing.T) {
	got := Count([]int{100})
	if got[100] != 1 || len(got) != 1 {
		t.Errorf("Count([100]) = %v, ожидалась карта с единственной парой 100:1", got)
	}
}
