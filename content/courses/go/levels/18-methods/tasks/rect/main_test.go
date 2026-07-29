package main

import "testing"

func TestRectArea(t *testing.T) {
	r := Rect{W: 3, H: 4}
	if got := r.Area(); got != 12 {
		t.Errorf("Rect{W: 3, H: 4}.Area() = %d, ожидалось 12", got)
	}
}

func TestRectPerimeter(t *testing.T) {
	r := Rect{W: 3, H: 4}
	if got := r.Perimeter(); got != 14 {
		t.Errorf("Rect{W: 3, H: 4}.Perimeter() = %d, ожидалось 14 — периметр это 2*(W+H)", got)
	}
}

func TestRectSquare(t *testing.T) {
	s := Rect{W: 5, H: 5}
	if got := s.Area(); got != 25 {
		t.Errorf("Rect{W: 5, H: 5}.Area() = %d, ожидалось 25 — квадрат тоже прямоугольник", got)
	}
	if got := s.Perimeter(); got != 20 {
		t.Errorf("Rect{W: 5, H: 5}.Perimeter() = %d, ожидалось 20", got)
	}
}

func TestRectZeroSide(t *testing.T) {
	z := Rect{W: 0, H: 7}
	if got := z.Area(); got != 0 {
		t.Errorf("Rect{W: 0, H: 7}.Area() = %d, ожидалось 0 — нулевая сторона обнуляет площадь", got)
	}
	if got := z.Perimeter(); got != 14 {
		t.Errorf("Rect{W: 0, H: 7}.Perimeter() = %d, ожидалось 14 — нулевая сторона не отменяет вторую пару сторон", got)
	}
}
