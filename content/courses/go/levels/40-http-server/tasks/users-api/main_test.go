package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doRequest(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPing(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodGet, "/ping")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /ping: статус = %d, ожидалось 200", rec.Code)
	}
	if got := strings.TrimSuffix(rec.Body.String(), "\n"); got != "pong" {
		t.Errorf("GET /ping: тело = %q, ожидалось %q", got, "pong")
	}
}

func TestPingWrongMethod(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodPost, "/ping")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /ping: статус = %d, ожидалось 405 — укажите метод прямо в шаблоне: \"GET /ping\"", rec.Code)
	}
}

func TestUserFound(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodGet, "/users/1")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /users/1: статус = %d, ожидалось 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("GET /users/1: Content-Type = %q, ожидался application/json — ставьте заголовок ДО записи тела", ct)
	}
	var u User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("GET /users/1: тело %q не разбирается как JSON: %v", rec.Body.String(), err)
	}
	if u != (User{ID: 1, Name: "Ада"}) {
		t.Errorf("GET /users/1: получено %+v, ожидалось {ID:1 Name:Ада}", u)
	}
}

func TestEveryUser(t *testing.T) {
	h := NewServer()
	for id, want := range users {
		rec := doRequest(t, h, http.MethodGet, fmt.Sprintf("/users/%d", id))
		if rec.Code != http.StatusOK {
			t.Errorf("GET /users/%d: статус = %d, ожидалось 200", id, rec.Code)
			continue
		}
		var u User
		if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
			t.Errorf("GET /users/%d: тело %q не разбирается как JSON: %v", id, rec.Body.String(), err)
			continue
		}
		if u != want {
			t.Errorf("GET /users/%d: получено %+v, ожидалось %+v", id, u, want)
		}
	}
}

func TestUserNotFound(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodGet, "/users/99")
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /users/99: статус = %d, ожидалось 404 — числа 99 нет в мапе users", rec.Code)
	}
}

func TestUserBadID(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodGet, "/users/abc")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GET /users/abc: статус = %d, ожидалось 400 — id не превращается в число", rec.Code)
	}
}

func TestUnknownPath(t *testing.T) {
	rec := doRequest(t, NewServer(), http.MethodGet, "/nope")
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /nope: статус = %d, ожидалось 404 — незнакомые пути отклоняет сам mux", rec.Code)
	}
}

func TestPingViaRealServer(t *testing.T) {
	srv := httptest.NewServer(NewServer())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ping")
	if err != nil {
		t.Fatalf("GET к настоящему тестовому серверу: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("статус через настоящий сервер = %d, ожидалось 200", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("чтение тела ответа: %v", err)
	}
	if got := strings.TrimSuffix(string(b), "\n"); got != "pong" {
		t.Errorf("тело через настоящий сервер = %q, ожидалось %q", got, "pong")
	}
}
