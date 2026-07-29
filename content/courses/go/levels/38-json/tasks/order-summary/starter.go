package main

import (
	"encoding/json"
	"os"
)

type Item struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
	Qty   int    `json:"qty"`
}

type Order struct {
	Customer string `json:"customer"`
	Items    []Item `json:"items"`
}

func main() {
	var o Order
	dec := json.NewDecoder(os.Stdin)
	// TODO: декодируйте заказ (указатель!) и напечатайте сводку:
	// <имя>: позиций <N>, на сумму <X> руб.
	_ = dec
	_ = o
}
