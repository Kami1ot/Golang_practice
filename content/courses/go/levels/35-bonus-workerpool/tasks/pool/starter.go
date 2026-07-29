package main

// Process применяет f к каждому элементу силами ровно workers горутин
// и возвращает результаты в исходном порядке (res[i] == f(nums[i])).
func Process(nums []int, workers int, f func(int) int) []int {
	res := make([]int, len(nums))
	// TODO: канал индексов + K воркеров + WaitGroup (import "sync")
	for i, n := range nums {
		res[i] = f(n) // пока последовательно — переделайте на пул
	}
	return res
}
