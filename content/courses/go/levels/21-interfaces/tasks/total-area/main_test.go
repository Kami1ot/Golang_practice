package main

import (
	"math"
	"testing"
)

func TestTotalAreaRects(t *testing.T) {
	shapes := []Shape{Rect{W: 3, H: 4}, Rect{W: 1, H: 2}}
	if got := TotalArea(shapes); math.Abs(got-14) > 1e-9 {
		t.Errorf("TotalArea([Rect{3,4}, Rect{1,2}]) = %v, ожидалось 14", got)
	}
}

func TestTotalAreaMixed(t *testing.T) {
	shapes := []Shape{Rect{W: 2, H: 3}, Circle{R: 1}, Rect{W: 1, H: 1}}
	want := 7 + math.Pi
	if got := TotalArea(shapes); math.Abs(got-want) > 1e-9 {
		t.Errorf("TotalArea для смеси прямоугольников и круга = %v, ожидалось %v — функция не должна различать конкретные типы", got, want)
	}
}

func TestTotalAreaSingle(t *testing.T) {
	if got := TotalArea([]Shape{Circle{R: 2}}); math.Abs(got-4*math.Pi) > 1e-9 {
		t.Errorf("TotalArea([Circle{2}]) = %v, ожидалось %v", got, 4*math.Pi)
	}
}

func TestTotalAreaEmpty(t *testing.T) {
	if got := TotalArea([]Shape{}); got != 0 {
		t.Errorf("TotalArea пустого слайса = %v, ожидалось 0", got)
	}
	if got := TotalArea(nil); got != 0 {
		t.Errorf("TotalArea(nil) = %v, ожидалось 0", got)
	}
}
