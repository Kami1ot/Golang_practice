package main

import "fmt"

func main() {
	var name string
	var count int
	var price float64
	fmt.Scan(&name, &count, &price)
	// TODO: выведите ценник в формате: НАЗВАНИЕ xКОЛИЧЕСТВО = СУММА руб.
	// Сумма — с двумя знаками после точки.
	_, _, _ = name, count, price
}
