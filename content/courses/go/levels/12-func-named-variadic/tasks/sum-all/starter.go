package main

import "fmt"

// Sum складывает любое число целых аргументов.
func Sum(nums ...int) int {
	// TODO: пройдитесь по nums циклом и верните сумму
	return 0
}

func main() {
	var n int
	fmt.Scan(&n)

	nums := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&nums[i])
	}

	// TODO: вызовите Sum с распаковкой слайса и напечатайте результат
	fmt.Println(0)
}
