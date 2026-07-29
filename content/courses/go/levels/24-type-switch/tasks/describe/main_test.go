package main

import "testing"

func TestDescribeInt(t *testing.T) {
	if got := Describe(42); got != "целое: 42" {
		t.Errorf("Describe(42) = %q, ожидалось %q", got, "целое: 42")
	}
	if got := Describe(-7); got != "целое: -7" {
		t.Errorf("Describe(-7) = %q, ожидалось %q", got, "целое: -7")
	}
}

func TestDescribeString(t *testing.T) {
	if got := Describe("го"); got != "строка: го" {
		t.Errorf("Describe(\"го\") = %q, ожидалось %q", got, "строка: го")
	}
	if got := Describe(""); got != "строка: " {
		t.Errorf("Describe(\"\") = %q, ожидалось %q — пустая строка всё ещё строка", got, "строка: ")
	}
}

func TestDescribeBool(t *testing.T) {
	if got := Describe(true); got != "булево: true" {
		t.Errorf("Describe(true) = %q, ожидалось %q", got, "булево: true")
	}
	if got := Describe(false); got != "булево: false" {
		t.Errorf("Describe(false) = %q, ожидалось %q", got, "булево: false")
	}
}

func TestDescribeUnknown(t *testing.T) {
	for _, v := range []any{3.14, []int{1, 2}, nil, struct{}{}} {
		if got := Describe(v); got != "неизвестный тип" {
			t.Errorf("Describe(%#v) = %q, ожидалось %q", v, got, "неизвестный тип")
		}
	}
}
