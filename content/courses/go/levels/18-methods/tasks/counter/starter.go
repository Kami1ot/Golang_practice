package main

// Counter — счётчик с внутренним состоянием.
// Не меняйте объявление типа.
type Counter struct {
	value int
}

// Inc увеличивает счётчик на 1.
func (c *Counter) Inc() {
	// TODO: увеличьте c.value
}

// Add изменяет счётчик на n (отрицательный n уменьшает).
func (c *Counter) Add(n int) {
	// TODO: прибавьте n к c.value
}

// Value возвращает текущее значение счётчика.
func (c Counter) Value() int {
	// TODO: верните c.value
	return 0
}
