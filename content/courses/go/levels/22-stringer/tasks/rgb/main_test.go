package main

import (
	"fmt"
	"testing"
)

// Компилируется только если RGB реализует fmt.Stringer.
var _ fmt.Stringer = RGB{}

func TestStringBasic(t *testing.T) {
	c := RGB{R: 255, G: 128, B: 0}
	if got := c.String(); got != "rgb(255, 128, 0)" {
		t.Errorf("RGB{255, 128, 0}.String() = %q, ожидалось %q", got, "rgb(255, 128, 0)")
	}
}

func TestStringZero(t *testing.T) {
	c := RGB{}
	if got := c.String(); got != "rgb(0, 0, 0)" {
		t.Errorf("RGB{}.String() = %q, ожидалось %q — нулевой цвет тоже цвет", got, "rgb(0, 0, 0)")
	}
}

func TestFmtUsesStringer(t *testing.T) {
	c := RGB{R: 1, G: 2, B: 3}
	if got := fmt.Sprint(c); got != "rgb(1, 2, 3)" {
		t.Errorf("fmt.Sprint(RGB{1, 2, 3}) = %q, ожидалось %q — fmt должен печатать через ваш String()", got, "rgb(1, 2, 3)")
	}
	if got := fmt.Sprintf("цвет: %v", c); got != "цвет: rgb(1, 2, 3)" {
		t.Errorf("Sprintf(%%v) = %q, ожидалось %q", got, "цвет: rgb(1, 2, 3)")
	}
}
