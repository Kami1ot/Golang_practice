package main

import (
	"bytes"
	"testing"
)

func TestGreeting(t *testing.T) {
	var buf bytes.Buffer
	WriteGreeting(&buf, "Гоша")
	if got := buf.String(); got != "Привет, Гоша!" {
		t.Errorf("WriteGreeting(w, \"Гоша\") записала %q, ожидалось %q", got, "Привет, Гоша!")
	}
}

func TestGreetingOther(t *testing.T) {
	var buf bytes.Buffer
	WriteGreeting(&buf, "мир")
	if got := buf.String(); got != "Привет, мир!" {
		t.Errorf("WriteGreeting(w, \"мир\") записала %q, ожидалось %q", got, "Привет, мир!")
	}
}

func TestGreetingAppends(t *testing.T) {
	var buf bytes.Buffer
	WriteGreeting(&buf, "раз")
	WriteGreeting(&buf, "два")
	want := "Привет, раз!Привет, два!"
	if got := buf.String(); got != want {
		t.Errorf("после двух вызовов в буфере %q, ожидалось %q — функция не должна добавлять перевод строки", got, want)
	}
}
