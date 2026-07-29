package main

import (
	"strings"
	"testing"
)

func TestProtectCalm(t *testing.T) {
	called := false
	err := Protect(func() { called = true })
	if err != nil {
		t.Errorf("Protect спокойной функции = %v, ожидался nil", err)
	}
	if !called {
		t.Errorf("Protect не вызвала переданную функцию")
	}
}

func TestProtectPanicString(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("паника вырвалась из Protect: %v", r)
		}
	}()
	err := Protect(func() { panic("взрыв") })
	if err == nil {
		t.Fatalf("Protect паникующей функции вернула nil, ожидалась ошибка")
	}
	if err.Error() != "паника: взрыв" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "паника: взрыв")
	}
}

func TestProtectPanicInt(t *testing.T) {
	err := Protect(func() { panic(42) })
	if err == nil {
		t.Fatalf("Protect функции с panic(42) вернула nil, ожидалась ошибка")
	}
	if err.Error() != "паника: 42" {
		t.Errorf("текст ошибки %q, ожидалось %q — %%v печатает любое значение паники", err.Error(), "паника: 42")
	}
}

func TestProtectRuntimePanic(t *testing.T) {
	err := Protect(func() {
		var m map[string]int
		m["x"] = 1
	})
	if err == nil {
		t.Fatalf("Protect функции с записью в nil map вернула nil, ожидалась ошибка")
	}
	if !strings.Contains(err.Error(), "nil map") {
		t.Errorf("текст ошибки %q должен содержать исходное сообщение runtime-паники (nil map)", err.Error())
	}
}

func TestProtectReusable(t *testing.T) {
	Protect(func() { panic("раз") })
	if err := Protect(func() {}); err != nil {
		t.Errorf("после пойманной паники Protect спокойной функции = %v, ожидался nil", err)
	}
}
