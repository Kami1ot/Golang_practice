package main

import (
	"math"
	"testing"
)

// Компилируется только если оба типа реализуют Shape.
var (
	_ Shape = Rect{}
	_ Shape = Circle{}
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestRectArea(t *testing.T) {
	r := Rect{W: 3, H: 4}
	if got := r.Area(); !almostEqual(got, 12) {
		t.Errorf("Rect{W: 3, H: 4}.Area() = %v, ожидалось 12", got)
	}
	r2 := Rect{W: 2.5, H: 4}
	if got := r2.Area(); !almostEqual(got, 10) {
		t.Errorf("Rect{W: 2.5, H: 4}.Area() = %v, ожидалось 10 — стороны бывают дробными", got)
	}
}

func TestCircleArea(t *testing.T) {
	c := Circle{R: 1}
	if got := c.Area(); !almostEqual(got, math.Pi) {
		t.Errorf("Circle{R: 1}.Area() = %v, ожидалось %v (число π)", got, math.Pi)
	}
	c2 := Circle{R: 2}
	if got := c2.Area(); !almostEqual(got, 4*math.Pi) {
		t.Errorf("Circle{R: 2}.Area() = %v, ожидалось %v (π·R² при R=2)", got, 4*math.Pi)
	}
}

func TestZeroValues(t *testing.T) {
	if got := (Rect{W: 0, H: 5}).Area(); !almostEqual(got, 0) {
		t.Errorf("Rect{W: 0, H: 5}.Area() = %v, ожидалось 0 — нулевая сторона обнуляет площадь", got)
	}
	if got := (Circle{R: 0}).Area(); !almostEqual(got, 0) {
		t.Errorf("Circle{R: 0}.Area() = %v, ожидалось 0 — у точки нет площади", got)
	}
}

func TestPolymorphism(t *testing.T) {
	shapes := []Shape{Rect{W: 2, H: 3}, Circle{R: 1}}
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	if want := 6 + math.Pi; !almostEqual(total, want) {
		t.Errorf("сумма площадей []Shape{Rect{2,3}, Circle{1}} = %v, ожидалось %v — оба типа должны работать через общий интерфейс", total, want)
	}
}
