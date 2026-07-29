package main

// Point — точка на плоскости.
// НЕ меняйте объявление: тесты используют именно этот тип.
type Point struct {
	X int
	Y int
}

// Move сдвигает точку: прибавляет dx к X и dy к Y.
func Move(p *Point, dx, dy int) {
	// TODO: измените координаты точки по указателю
}

// Reset возвращает точку в начало координат: обнуляет X и Y.
func Reset(p *Point) {
	// TODO: обнулите обе координаты
}
