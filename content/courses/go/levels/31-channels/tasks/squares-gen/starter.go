package main

// Squares возвращает канал, в который горутина отправляет
// квадраты 1..n по порядку, а затем закрывает его.
func Squares(n int) <-chan int {
	ch := make(chan int)
	// TODO: горутина с циклом отправок и close после него
	close(ch)
	return ch
}
