package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func callHello(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	HelloHandler(rec, req)
	return rec
}

func helloBody(rec *httptest.ResponseRecorder) string {
	return strings.TrimSuffix(rec.Body.String(), "\n")
}

func TestHelloWithName(t *testing.T) {
	rec := callHello(t, "/hello?name=Gopher")
	if rec.Code != http.StatusOK {
		t.Errorf("статус = %d, ожидалось 200 — WriteHeader здесь трогать не нужно", rec.Code)
	}
	if got := helloBody(rec); got != "Привет, Gopher!" {
		t.Errorf("тело для ?name=Gopher = %q, ожидалось %q", got, "Привет, Gopher!")
	}
}

func TestHelloWithoutName(t *testing.T) {
	rec := callHello(t, "/hello")
	if got := helloBody(rec); got != "Привет, гость!" {
		t.Errorf("тело без параметра = %q, ожидалось %q — нет name, значит гость", got, "Привет, гость!")
	}
}

func TestHelloEmptyName(t *testing.T) {
	rec := callHello(t, "/hello?name=")
	if got := helloBody(rec); got != "Привет, гость!" {
		t.Errorf("тело для пустого ?name= = %q, ожидалось %q — пустое имя считается отсутствующим", got, "Привет, гость!")
	}
}

func TestHelloCyrillic(t *testing.T) {
	rec := callHello(t, "/hello?name="+url.QueryEscape("Ада Лавлейс"))
	if got := helloBody(rec); got != "Привет, Ада Лавлейс!" {
		t.Errorf("тело = %q, ожидалось %q — кириллица приезжает в URL-кодировании, Query().Get сам её раскодирует", got, "Привет, Ада Лавлейс!")
	}
}
