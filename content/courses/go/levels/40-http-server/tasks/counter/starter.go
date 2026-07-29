package main

import (
	"net/http"
	"sync"
)

// addRequest — тело запроса POST /add: {"n": 5}.
type addRequest struct {
	N int `json:"n"`
}

// totalResponse — тело ответа GET /total: {"total": 12}.
type totalResponse struct {
	Total int `json:"total"`
}

// counter — состояние сервиса: сумма под защитой мьютекса.
type counter struct {
	mu    sync.Mutex
	total int
}

// NewCounter собирает счётчик:
//
//	POST /add   — JSON {"n": 5} прибавляет n (кривой JSON → 400)
//	GET  /total — JSON {"total": ...} с Content-Type: application/json
//
// Каждый вызов NewCounter — независимый счётчик с нуля.
func NewCounter() http.Handler {
	c := &counter{}
	_ = c // уберите, когда хендлеры начнут использовать c
	mux := http.NewServeMux()
	// TODO: mux.HandleFunc("POST /add", ...) — json.NewDecoder(r.Body).Decode
	//       в addRequest; ошибка → http.Error(..., 400) и return; сумму менять под c.mu
	// TODO: mux.HandleFunc("GET /total", ...) — прочитать сумму под c.mu,
	//       затем Content-Type и json.NewEncoder(w).Encode(totalResponse{...})
	//       (понадобится import encoding/json)
	return mux
}
