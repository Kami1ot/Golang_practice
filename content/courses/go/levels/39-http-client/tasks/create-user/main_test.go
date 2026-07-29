package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type gotRequest struct {
	method      string
	path        string
	contentType string
	body        []byte
}

func newCreateServer(reqs chan<- gotRequest) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		select {
		case reqs <- gotRequest{r.Method, r.URL.Path, r.Header.Get("Content-Type"), body}:
		default:
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":101}`)
	}))
}

func TestCreateUserReturnsID(t *testing.T) {
	reqs := make(chan gotRequest, 1)
	srv := newCreateServer(reqs)
	defer srv.Close()

	id, err := CreateUser(srv.URL, User{Name: "Гоша", Email: "gosha@go.dev"})
	if err != nil {
		t.Fatalf("CreateUser вернула ошибку: %v", err)
	}
	if id != 101 {
		t.Errorf("CreateUser = %d, ожидалось 101 — id нужно взять из JSON-ответа сервера", id)
	}
}

func TestCreateUserSendsProperRequest(t *testing.T) {
	reqs := make(chan gotRequest, 1)
	srv := newCreateServer(reqs)
	defer srv.Close()

	_, err := CreateUser(srv.URL, User{Name: "Ада", Email: "ada@go.dev"})
	if err != nil {
		t.Fatalf("CreateUser вернула ошибку: %v", err)
	}

	var req gotRequest
	select {
	case req = <-reqs:
	case <-time.After(2 * time.Second):
		t.Fatalf("на сервер не пришло ни одного запроса — CreateUser должна делать POST на baseURL+\"/users\"")
	}
	if req.method != http.MethodPost {
		t.Errorf("метод запроса %q, ожидался POST", req.method)
	}
	if req.path != "/users" {
		t.Errorf("путь запроса %q, ожидался /users", req.path)
	}
	if !strings.HasPrefix(req.contentType, "application/json") {
		t.Errorf("Content-Type запроса %q, ожидался application/json — это второй аргумент http.Post", req.contentType)
	}
	var sent User
	if err := json.Unmarshal(req.body, &sent); err != nil {
		t.Fatalf("тело запроса — не валидный JSON: %v (получено %q)", err, req.body)
	}
	want := User{Name: "Ада", Email: "ada@go.dev"}
	if sent != want {
		t.Errorf("на сервер уехало %+v, ожидалось %+v — сериализуйте именно переданного пользователя", sent, want)
	}
}

func TestCreateUserBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	id, err := CreateUser(srv.URL, User{Name: "Гоша"})
	if err == nil {
		t.Fatalf("CreateUser при ответе 500 вернула nil-ошибку — статус нужно проверять")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("в тексте ошибки нет кода статуса: %q", err)
	}
	if id != 0 {
		t.Errorf("при ошибке CreateUser должна возвращать 0, получено %d", id)
	}
}

func TestCreateUserStatus200IsNotCreated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":5}`) // статус по умолчанию 200 — но успех создания это 201
	}))
	defer srv.Close()

	_, err := CreateUser(srv.URL, User{Name: "Гоша"})
	if err == nil {
		t.Fatalf("сервер ответил 200 вместо 201, а CreateUser не вернула ошибку — сравнивайте именно с http.StatusCreated")
	}
	if !strings.Contains(err.Error(), "200") {
		t.Errorf("в тексте ошибки нет кода статуса: %q", err)
	}
}
