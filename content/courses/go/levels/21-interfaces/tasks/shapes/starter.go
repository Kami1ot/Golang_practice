package main

import "math"

// Shape — контракт фигуры: любой тип с методом Area() float64.
// Не меняйте объявление интерфейса.
type Shape interface {
	Area() float64
}

// Rect — прямоугольник со сторонами W и H.
type Rect struct {
	W, H float64
}

// Circle — круг с радиусом R.
type Circle struct {
	R float64
}

// Area возвращает площадь прямоугольника.
func (r Rect) Area() float64 {
	// TODO: перемножьте стороны
	return 0
}

// Area возвращает площадь круга: π·R².
func (c Circle) Area() float64 {
	// TODO: замените заглушку на формулу площади круга
	return math.Pi * 0
}
