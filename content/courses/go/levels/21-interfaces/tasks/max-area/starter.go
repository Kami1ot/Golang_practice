package main

import "math"

// Всё, что ниже, уже готово — менять не нужно.
type Shape interface {
	Area() float64
}

type Rect struct {
	W, H float64
}

func (r Rect) Area() float64 {
	return r.W * r.H
}

type Circle struct {
	R float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

// MaxArea возвращает фигуру с наибольшей площадью и true.
// Для пустого слайса — nil и false.
// Если максимумов несколько — первая из них.
func MaxArea(shapes []Shape) (Shape, bool) {
	// TODO: пустой слайс → nil, false; иначе выберите лидера циклом
	return nil, false
}
