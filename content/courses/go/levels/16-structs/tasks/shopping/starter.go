package main

import "fmt"

type Item struct {
	Name  string
	Price int
}

func main() {
	var n int
	fmt.Scan(&n)
	// TODO: прочитайте n товаров в слайс []Item, затем выведите
	// общую сумму и название самого дорогого (при ничьей — первого)
}
