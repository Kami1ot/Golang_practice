package main

import (
	"fmt"
	"io"
)

// Item — товар в чеке: название и цена в рублях.
// Не меняйте объявление типа.
type Item struct {
	Name  string
	Price int
}

// WriteReceipt пишет в w строку на каждый товар ("- хлеб: 40 ₽")
// и итоговую строку ("итого: 130 ₽"). Каждая строка заканчивается \n.
func WriteReceipt(w io.Writer, items []Item) {
	// TODO: цикл по товарам + итоговая строка
	fmt.Fprint(w, "")
}
