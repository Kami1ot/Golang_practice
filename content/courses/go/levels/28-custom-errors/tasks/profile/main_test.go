package main

import (
	"errors"
	"testing"
)

var _ error = NotFoundError{}

func base() map[string]string {
	return map[string]string{"anna": "active", "vasya": "banned"}
}

func TestOpenProfileOK(t *testing.T) {
	if err := OpenProfile(base(), "anna"); err != nil {
		t.Errorf("OpenProfile для активного профиля = %v, ожидался nil", err)
	}
}

func TestOpenProfileEmptyName(t *testing.T) {
	err := OpenProfile(base(), "")
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("для пустого имени ожидался ErrEmptyName, получено %v", err)
	}
}

func TestOpenProfileNotFound(t *testing.T) {
	err := OpenProfile(base(), "kot")
	if err == nil {
		t.Fatalf("для неизвестного имени ожидалась ошибка, получен nil")
	}
	var nf NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("errors.As не нашла NotFoundError, получено %v", err)
	}
	if nf.Name != "kot" {
		t.Errorf("NotFoundError.Name = %q, ожидалось %q", nf.Name, "kot")
	}
	if got, want := nf.Error(), `профиль "kot" не найден`; got != want {
		t.Errorf("текст NotFoundError = %q, ожидалось %q", got, want)
	}
}

func TestOpenProfileBanned(t *testing.T) {
	err := OpenProfile(base(), "vasya")
	if err == nil {
		t.Fatalf("для заблокированного ожидалась ошибка, получен nil")
	}
	if !errors.Is(err, ErrBanned) {
		t.Errorf("errors.Is(err, ErrBanned) = false — оборачивайте sentinel глаголом %%w, получено: %v", err)
	}
	if got, want := err.Error(), "доступ к vasya: пользователь заблокирован"; got != want {
		t.Errorf("текст ошибки %q, ожидалось %q", got, want)
	}
}

func TestOpenProfileOrder(t *testing.T) {
	// пустое имя важнее «не найдено»: guard должен идти первым
	err := OpenProfile(map[string]string{}, "")
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("для пустого имени на пустой карте ожидался ErrEmptyName, получено %v", err)
	}
}
