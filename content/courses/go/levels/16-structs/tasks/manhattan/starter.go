package main

import "fmt"

type Point struct {
	X, Y int
}

// Dist возвращает манхэттенское расстояние между точками a и b.
func Dist(a, b Point) int {
	// TODO: |a.X-b.X| + |a.Y-b.Y| — модуль своей функцией abs
	return 0
}

func main() {
	var x1, y1, x2, y2 int
	fmt.Scan(&x1, &y1, &x2, &y2)
	// TODO: соберите две точки Point и напечатайте Dist для них
}
