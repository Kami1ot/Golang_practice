package main

import "fmt"

// Describe опознаёт тип значения: "целое: 42", "строка: привет",
// "булево: true" или "неизвестный тип".
func Describe(v any) string {
	// TODO: switch t := v.(type) с ветками int, string, bool и default
	return fmt.Sprint("")
}
