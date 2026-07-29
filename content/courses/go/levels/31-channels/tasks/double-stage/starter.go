package main

// DoubleAll — стадия конвейера: читает из in, удваивает, шлёт в выход.
// После исчерпания входа закрывает выходной канал.
func DoubleAll(in <-chan int) <-chan int {
	out := make(chan int)
	// TODO: горутина: range по in → отправка v*2 → close(out)
	close(out)
	return out
}
