package main

import (
	"encoding/json"
	"os"
)

type Item struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func main() {
	var items []Item
	dec := json.NewDecoder(os.Stdin)
	// TODO: декодируйте JSON-массив в items (указатель!), затем
	// выведите сумму цен и имя самого дешёвого товара
	_ = dec
	_ = items
}
