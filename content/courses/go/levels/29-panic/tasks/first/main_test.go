package main

import "testing"

func TestFirstOK(t *testing.T) {
	got, err := First([]int{7, 8, 9})
	if err != nil {
		t.Fatalf("First([7, 8, 9]) вернула ошибку %v, ожидался nil", err)
	}
	if got != 7 {
		t.Errorf("First([7, 8, 9]) = %d, ожидалось 7", got)
	}
}

func TestFirstSingle(t *testing.T) {
	got, err := First([]int{42})
	if err != nil {
		t.Fatalf("First([42]) вернула ошибку %v, ожидался nil", err)
	}
	if got != 42 {
		t.Errorf("First([42]) = %d, ожидалось 42", got)
	}
}

func TestFirstEmptyNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("First пустого слайса ЗАПАНИКОВАЛА: %v — нужна guard-проверка длины", r)
		}
	}()
	got, err := First([]int{})
	if err == nil {
		t.Fatalf("First([]) вернула nil, ожидалась ошибка")
	}
	if err.Error() != "пустой слайс" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "пустой слайс")
	}
	if got != 0 {
		t.Errorf("при ошибке результат = %d, ожидалось 0", got)
	}
}

func TestFirstNilNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("First(nil) ЗАПАНИКОВАЛА: %v", r)
		}
	}()
	if _, err := First(nil); err == nil {
		t.Errorf("First(nil) вернула nil, ожидалась ошибка")
	}
}
