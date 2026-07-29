package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "привет, клиент!")
	}))
	defer srv.Close()

	got, err := Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch при ответе 200 вернула ошибку: %v", err)
	}
	if got != "привет, клиент!" {
		t.Errorf("Fetch = %q, ожидалось %q", got, "привет, клиент!")
	}
}

func TestFetchNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r) // отвечает 404
	}))
	defer srv.Close()

	got, err := Fetch(srv.URL)
	if err == nil {
		t.Fatalf("Fetch при ответе 404 вернула nil-ошибку — для http.Get 404 не ошибка, статус проверяем сами")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("в тексте ошибки нет кода статуса: %q — подставьте resp.StatusCode через %%d", err)
	}
	if got != "" {
		t.Errorf("при ошибке Fetch должна возвращать пустую строку, получено %q", got)
	}
}

func TestFetchServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := Fetch(srv.URL)
	if err == nil {
		t.Fatalf("Fetch при ответе 500 вернула nil-ошибку — любой статус, кроме 200, должен превращаться в ошибку")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("в тексте ошибки нет кода статуса: %q", err)
	}
}

func TestFetchLargeBody(t *testing.T) {
	big := strings.Repeat("go!", 50000) // 150 000 байт — за один Read столько не приезжает
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, big)
	}))
	defer srv.Close()

	got, err := Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch большого тела вернула ошибку: %v", err)
	}
	if got != big {
		t.Errorf("Fetch вернула %d байт из %d — тело нужно дочитывать до конца (io.ReadAll)", len(got), len(big))
	}
}

func TestFetchUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // сервер уже выключен — доставка невозможна

	_, err := Fetch(url)
	if err == nil {
		t.Fatalf("Fetch по адресу выключенного сервера вернула nil-ошибку — ошибку http.Get нужно возвращать, а не глотать")
	}
}
