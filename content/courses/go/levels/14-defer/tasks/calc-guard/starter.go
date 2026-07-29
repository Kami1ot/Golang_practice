package main

import "fmt"

func main() {
	// TODO: самой первой строкой отложите печать "готово"

	var a, b int
	var op string
	fmt.Scan(&a, &op, &b)

	// TODO: разберите знак операции switch'ем; деление на ноль
	// и неизвестный знак — ошибки (тексты в условии)
	_, _, _ = a, b, op
}
