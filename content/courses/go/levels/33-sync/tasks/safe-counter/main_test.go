package main

import (
	"sync"
	"testing"
)

func TestSafeCounterSequential(t *testing.T) {
	c := SafeCounter{}
	for i := 0; i < 5; i++ {
		c.Inc()
	}
	if got := c.Value(); got != 5 {
		t.Errorf("после 5 Inc подряд Value() = %d, ожидалось 5", got)
	}
}

func TestSafeCounterConcurrent(t *testing.T) {
	c := SafeCounter{}
	var wg sync.WaitGroup
	const goroutines = 200
	const perGoroutine = 50
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()
	want := goroutines * perGoroutine
	if got := c.Value(); got != want {
		t.Errorf("после %d конкурентных Inc Value() = %d, ожидалось %d — обновления теряются без замка", want, got, want)
	}
}

func TestSafeCounterReadDuringWrites(t *testing.T) {
	c := SafeCounter{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
		go func() {
			defer wg.Done()
			_ = c.Value() // читаем параллельно с записью — Value тоже должен брать замок
		}()
	}
	wg.Wait()
	if got := c.Value(); got != 50 {
		t.Errorf("после 50 Inc с параллельными чтениями Value() = %d, ожидалось 50", got)
	}
}
