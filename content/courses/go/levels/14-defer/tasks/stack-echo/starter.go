package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	for i := 0; i < n; i++ {
		var w string
		fmt.Scan(&w)
		// TODO: напечатайте "-> слово" сразу,
		// а печать "<- слово" отложите через defer
		_ = w
	}
}
