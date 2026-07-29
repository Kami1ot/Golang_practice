package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	seen := make(map[int]bool)
	// TODO: прочитайте n чисел и отметьте каждое в карте

	fmt.Println(len(seen))
}
