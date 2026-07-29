package main

import (
	"encoding/json"
	"os"
)

type Book struct {
	Title string `json:"title"`
	Pages int    `json:"pages"`
}

func main() {
	var b Book
	dec := json.NewDecoder(os.Stdin)
	// TODO: декодируйте ввод в структуру b (не забудьте указатель!)
	// и выведите две строки: заголовок и число страниц
	_ = dec
	_ = b
}
