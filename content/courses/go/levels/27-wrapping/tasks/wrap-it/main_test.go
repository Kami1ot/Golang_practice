package main

import (
	"errors"
	"testing"
)

func TestProcessOK(t *testing.T) {
	if err := Process("привет"); err != nil {
		t.Errorf("Process(\"привет\") = %v, ожидался nil", err)
	}
}

func TestProcessMessage(t *testing.T) {
	err := Process("")
	if err == nil {
		t.Fatalf("Process(\"\") = nil, ожидалась ошибка")
	}
	if err.Error() != "обработка: пустой ввод" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "обработка: пустой ввод")
	}
}

func TestProcessWraps(t *testing.T) {
	err := Process("")
	if err == nil {
		t.Fatalf("Process(\"\") = nil, ожидалась ошибка")
	}
	if !errors.Is(err, ErrEmpty) {
		t.Errorf("errors.Is(err, ErrEmpty) = false — исходная ошибка потеряна: оборачивайте глаголом %%w, а не %%v")
	}
	if err == ErrEmpty {
		t.Errorf("Process вернула ErrEmpty без обёртки — добавьте контекст \"обработка: \"")
	}
}
