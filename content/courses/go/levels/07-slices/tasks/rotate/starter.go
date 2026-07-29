package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	var nums []int
	for range n {
		var x int
		fmt.Scan(&x)
		nums = append(nums, x)
	}

	var k int
	fmt.Scan(&k)

	// TODO: соберите новый слайс — сначала nums[k:], затем nums[:k]
	_ = k
	fmt.Println(nums)
}
