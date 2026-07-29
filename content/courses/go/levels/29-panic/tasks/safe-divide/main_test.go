package main

import (
	"strings"
	"testing"
)

func TestSafeDivideOK(t *testing.T) {
	got, err := SafeDivide(10, 2)
	if err != nil {
		t.Fatalf("SafeDivide(10, 2) вернула ошибку %v, ожидался nil", err)
	}
	if got != 5 {
		t.Errorf("SafeDivide(10, 2) = %d, ожидалось 5", got)
	}
}

func TestSafeDivideNegative(t *testing.T) {
	got, err := SafeDivide(-8, 4)
	if err != nil {
		t.Fatalf("SafeDivide(-8, 4) вернула ошибку %v, ожидался nil", err)
	}
	if got != -2 {
		t.Errorf("SafeDivide(-8, 4) = %d, ожидалось -2", got)
	}
}

func TestSafeDivideZeroCaught(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("паника вырвалась наружу: %v — recover должен поймать её внутри SafeDivide", r)
		}
	}()
	got, err := SafeDivide(1, 0)
	if err == nil {
		t.Fatalf("SafeDivide(1, 0) вернула nil, ожидалась ошибка из recover")
	}
	if !strings.HasPrefix(err.Error(), "паника: ") {
		t.Errorf("текст ошибки %q должен начинаться с %q", err.Error(), "паника: ")
	}
	if !strings.Contains(err.Error(), "divide by zero") {
		t.Errorf("текст ошибки %q должен содержать исходное сообщение паники (divide by zero) — вставляйте r через %%v", err.Error())
	}
	if got != 0 {
		t.Errorf("при пойманной панике result = %d, ожидалось 0", got)
	}
}

func TestSafeDivideStillWorksAfterPanic(t *testing.T) {
	SafeDivide(5, 0)
	got, err := SafeDivide(9, 3)
	if err != nil || got != 3 {
		t.Errorf("после пойманной паники SafeDivide(9, 3) = (%d, %v), ожидалось (3, nil)", got, err)
	}
}
