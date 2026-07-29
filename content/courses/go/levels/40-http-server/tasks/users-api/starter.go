package main

import (
	"net/http"
)

// User — пользователь в ответе API.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// users — «база данных» сервера. Не меняйте её.
var users = map[int]User{
	1: {ID: 1, Name: "Ада"},
	2: {ID: 2, Name: "Линус"},
	3: {ID: 3, Name: "Роб"},
}

// NewServer собирает маршрутизатор API:
//
//	GET /ping       → "pong"
//	GET /users/{id} → JSON пользователя (400 — id не число, 404 — нет такого)
func NewServer() http.Handler {
	mux := http.NewServeMux()
	// TODO: mux.HandleFunc("GET /ping", ...) — просто написать pong в w
	// TODO: mux.HandleFunc("GET /users/{id}", ...) — strconv.Atoi(r.PathValue("id")),
	//       мапа users, Content-Type и json.NewEncoder(w).Encode
	//       (понадобятся import'ы fmt, strconv и encoding/json)
	return mux
}
