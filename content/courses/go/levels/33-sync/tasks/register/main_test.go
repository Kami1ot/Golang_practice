package main

import (
	"sync"
	"testing"
)

func TestRegisterBasic(t *testing.T) {
	r := NewRegister()
	r.Add("Аня", 10)
	r.Add("Аня", 5)
	r.Add("Боря", 7)
	if got := r.Total("Аня"); got != 15 {
		t.Errorf("Total(\"Аня\") = %d, ожидалось 15", got)
	}
	if got := r.Total("Боря"); got != 7 {
		t.Errorf("Total(\"Боря\") = %d, ожидалось 7", got)
	}
	if got := r.Total("Юра"); got != 0 {
		t.Errorf("Total незнакомого игрока = %d, ожидалось 0", got)
	}
}

func TestRegisterNoNilMapPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Add запаниковала: %v — NewRegister должен создать карту", r)
		}
	}()
	r := NewRegister()
	r.Add("первый", 1)
}

func TestRegisterConcurrentOneName(t *testing.T) {
	r := NewRegister()
	var wg sync.WaitGroup
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Add("Аня", 2)
		}()
	}
	wg.Wait()
	if got := r.Total("Аня"); got != 600 {
		t.Errorf("300 конкурентных Add по 2 очка: Total = %d, ожидалось 600 — очки теряются без замка", got)
	}
}

func TestRegisterConcurrentManyNames(t *testing.T) {
	r := NewRegister()
	names := []string{"а", "б", "в", "г"}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		for _, name := range names {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				r.Add(name, 1)
				_ = r.Total(name) // читаем параллельно с записями
			}(name)
		}
	}
	wg.Wait()
	for _, name := range names {
		if got := r.Total(name); got != 100 {
			t.Errorf("Total(%q) = %d, ожидалось 100", name, got)
		}
	}
}
