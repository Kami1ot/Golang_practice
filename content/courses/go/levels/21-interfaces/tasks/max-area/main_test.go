package main

import (
	"math"
	"testing"
)

func TestMaxAreaMixed(t *testing.T) {
	shapes := []Shape{Rect{W: 1, H: 1}, Circle{R: 2}, Rect{W: 3, H: 4}}
	got, ok := MaxArea(shapes)
	if !ok {
		t.Fatalf("MaxArea непустого слайса вернула ok=false, ожидалось true")
	}
	if got != (Circle{R: 2}) {
		t.Errorf("MaxArea вернула %#v, ожидался Circle{R: 2} (площадь ≈12.57 против 12 у Rect{3,4})", got)
	}
}

func TestMaxAreaSingle(t *testing.T) {
	got, ok := MaxArea([]Shape{Rect{W: 2, H: 2}})
	if !ok {
		t.Fatalf("MaxArea слайса из одной фигуры вернула ok=false, ожидалось true")
	}
	if got != (Rect{W: 2, H: 2}) {
		t.Errorf("MaxArea из одной фигуры вернула %#v, ожидался сам Rect{W: 2, H: 2}", got)
	}
}

func TestMaxAreaFirstOfEqual(t *testing.T) {
	shapes := []Shape{Rect{W: 2, H: 6}, Rect{W: 6, H: 2}, Rect{W: 3, H: 4}}
	got, ok := MaxArea(shapes)
	if !ok {
		t.Fatalf("MaxArea непустого слайса вернула ok=false, ожидалось true")
	}
	if got != (Rect{W: 2, H: 6}) {
		t.Errorf("при равных площадях 12 ожидалась ПЕРВАЯ фигура Rect{W: 2, H: 6}, получено %#v — используйте строгое сравнение >", got)
	}
}

func TestMaxAreaEmpty(t *testing.T) {
	got, ok := MaxArea([]Shape{})
	if ok {
		t.Errorf("MaxArea пустого слайса вернула ok=true, ожидалось false")
	}
	if got != nil {
		t.Errorf("MaxArea пустого слайса вернула %#v, ожидался nil", got)
	}
	if _, ok := MaxArea(nil); ok {
		t.Errorf("MaxArea(nil) вернула ok=true, ожидалось false")
	}
}

func TestMaxAreaValue(t *testing.T) {
	shapes := []Shape{Circle{R: 1}, Circle{R: 3}, Circle{R: 2}}
	got, ok := MaxArea(shapes)
	if !ok {
		t.Fatalf("MaxArea непустого слайса вернула ok=false, ожидалось true")
	}
	if math.Abs(got.Area()-9*math.Pi) > 1e-9 {
		t.Errorf("площадь найденной фигуры = %v, ожидалось %v (круг радиуса 3)", got.Area(), 9*math.Pi)
	}
}
