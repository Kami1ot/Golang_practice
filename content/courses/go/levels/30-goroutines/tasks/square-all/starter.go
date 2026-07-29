package main

// SquareAll считает квадрат каждого элемента в отдельной горутине.
// Каждая горутина пишет только в свой слот результата; возврат — после Wait.
func SquareAll(nums []int) []int {
	res := make([]int, len(nums))
	// TODO: горутина на каждый индекс + WaitGroup (import "sync")
	for i, n := range nums {
		res[i] = n * n // пока последовательно — переделайте на горутины
	}
	return res
}
