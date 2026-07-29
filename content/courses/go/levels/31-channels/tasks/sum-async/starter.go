package main

// SumAsync сразу возвращает канал; сумму a+b считает горутина
// и отправляет её в этот канал.
func SumAsync(a, b int) <-chan int {
	// TODO: make → go func с отправкой → return
	ch := make(chan int)
	return ch
}
