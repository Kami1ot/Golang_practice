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

// TotalArea возвращает сумму площадей всех фигур слайса.
// Для пустого слайса — 0.
func TotalArea(shapes []Shape) float64 {
	// TODO: пройдитесь по слайсу range-циклом и сложите площади
	return 0
}
