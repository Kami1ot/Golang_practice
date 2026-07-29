package main

import "sync"

// SafeCounter — потокобезопасный счётчик. Не меняйте объявление.
type SafeCounter struct {
	mu sync.Mutex
	n  int
}

// Inc увеличивает счётчик на 1 (безопасно для горутин).
func (c *SafeCounter) Inc() {
	// TODO: критическая секция вокруг c.n++
	c.n++
}

// Value возвращает текущее значение (тоже под замком!).
func (c *SafeCounter) Value() int {
	// TODO: критическая секция вокруг чтения
	return c.n
}
