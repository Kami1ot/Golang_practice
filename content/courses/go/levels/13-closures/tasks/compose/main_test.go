package main

import "testing"

func TestComposeOrder(t *testing.T) {
	inc := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }
	h := Compose(inc, double)
	if h == nil {
		t.Fatal("Compose вернул nil, ожидалась функция")
	}
	if got := h(3); got != 7 {
		t.Errorf("Compose(inc, double)(3) = %d, ожидалось 7: сначала g (3*2 = 6), потом f (6+1 = 7). Проверьте порядок: h(x) = f(g(x))", got)
	}
	h2 := Compose(double, inc)
	if got := h2(3); got != 8 {
		t.Errorf("Compose(double, inc)(3) = %d, ожидалось 8: сначала g (3+1 = 4), потом f (4*2 = 8). Проверьте порядок: h(x) = f(g(x))", got)
	}
}

func TestComposeWithItself(t *testing.T) {
	double := func(x int) int { return x * 2 }
	quad := Compose(double, double)
	if quad == nil {
		t.Fatal("Compose вернул nil, ожидалась функция")
	}
	if got := quad(3); got != 12 {
		t.Errorf("Compose(double, double)(3) = %d, ожидалось 12 — удвоение, применённое дважды", got)
	}
}

func TestComposeReusable(t *testing.T) {
	square := func(x int) int { return x * x }
	inc := func(x int) int { return x + 1 }
	h := Compose(square, inc)
	if h == nil {
		t.Fatal("Compose вернул nil, ожидалась функция")
	}
	if got := h(0); got != 1 {
		t.Errorf("h = Compose(square, inc); h(0) = %d, ожидалось 1 (сначала 0+1, потом квадрат)", got)
	}
	if got := h(1); got != 4 {
		t.Errorf("h = Compose(square, inc); h(1) = %d, ожидалось 4 — возвращённая функция должна работать при повторных вызовах", got)
	}
	if got := h(4); got != 25 {
		t.Errorf("h = Compose(square, inc); h(4) = %d, ожидалось 25 — возвращённая функция должна работать при повторных вызовах", got)
	}
}
