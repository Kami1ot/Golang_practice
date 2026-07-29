package main

// RunAll запускает каждую функцию слайса в отдельной горутине
// и возвращается после завершения всех. Понадобится sync.WaitGroup.
func RunAll(funcs []func()) {
	// TODO: Add до go, defer Done внутри, Wait после цикла
	for _, f := range funcs {
		f() // пока последовательно — переделайте на горутины
	}
}
