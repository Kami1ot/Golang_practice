package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newUsersServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/7":
			fmt.Fprint(w, `{"id":7,"name":"Гоша"}`)
		case "/users/42":
			fmt.Fprint(w, `{"id":42,"name":"Ада"}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestFetchUserOK(t *testing.T) {
	srv := newUsersServer()
	defer srv.Close()

	got, err := FetchUser(srv.URL, 7)
	if err != nil {
		t.Fatalf("FetchUser(7) вернула ошибку: %v — проверьте, что URL собран как baseURL+\"/users/7\"", err)
	}
	want := User{ID: 7, Name: "Гоша"}
	if got != want {
		t.Errorf("FetchUser(7) = %+v, ожидалось %+v", got, want)
	}
}

func TestFetchUserAnotherID(t *testing.T) {
	srv := newUsersServer()
	defer srv.Close()

	got, err := FetchUser(srv.URL, 42)
	if err != nil {
		t.Fatalf("FetchUser(42) вернула ошибку: %v — id должен подставляться в URL, а не быть зашитым", err)
	}
	want := User{ID: 42, Name: "Ада"}
	if got != want {
		t.Errorf("FetchUser(42) = %+v, ожидалось %+v", got, want)
	}
}

func TestFetchUserNotFound(t *testing.T) {
	srv := newUsersServer()
	defer srv.Close()

	got, err := FetchUser(srv.URL, 99)
	if err == nil {
		t.Fatalf("FetchUser(99) вернула nil-ошибку — сервер ответил 404, статус нужно проверять")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("в тексте ошибки нет кода статуса: %q", err)
	}
	if got != (User{}) {
		t.Errorf("при ошибке нужно возвращать нулевого User{}, получено %+v", got)
	}
}

func TestFetchUserBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "это точно не JSON")
	}))
	defer srv.Close()

	_, err := FetchUser(srv.URL, 1)
	if err == nil {
		t.Fatalf("FetchUser со сломанным JSON в ответе вернула nil-ошибку — ошибку декодера нужно возвращать")
	}
}
