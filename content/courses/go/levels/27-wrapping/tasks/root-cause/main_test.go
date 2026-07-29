package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestRootCauseDeep(t *testing.T) {
	base := errors.New("диск отвалился")
	wrapped := fmt.Errorf("сохранение: %w", fmt.Errorf("запись файла: %w", base))
	if got := RootCause(wrapped); got != base {
		t.Errorf("RootCause цепочки из трёх слоёв = %v, ожидалась исходная %v", got, base)
	}
}

func TestRootCauseSingleWrap(t *testing.T) {
	base := errors.New("корень")
	if got := RootCause(fmt.Errorf("слой: %w", base)); got != base {
		t.Errorf("RootCause одной обёртки = %v, ожидалась %v", got, base)
	}
}

func TestRootCauseNoWrap(t *testing.T) {
	base := errors.New("голая ошибка")
	if got := RootCause(base); got != base {
		t.Errorf("RootCause ошибки без обёрток = %v, ожидалась она сама", got)
	}
}

func TestRootCauseNil(t *testing.T) {
	if got := RootCause(nil); got != nil {
		t.Errorf("RootCause(nil) = %v, ожидался nil", got)
	}
}

func TestRootCauseMessage(t *testing.T) {
	base := errors.New("первопричина")
	deep := fmt.Errorf("a: %w", fmt.Errorf("b: %w", fmt.Errorf("c: %w", base)))
	got := RootCause(deep)
	if got == nil || got.Error() != "первопричина" {
		t.Errorf("RootCause глубокой цепочки = %v, ожидалась ошибка с текстом %q", got, "первопричина")
	}
}
