package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func postAdd(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func getTotal(t *testing.T, h http.Handler) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/total", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /total: статус = %d, ожидалось 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("GET /total: Content-Type = %q, ожидался application/json — ставьте заголовок ДО записи тела", ct)
	}
	var out struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("GET /total: тело %q не разбирается как JSON вида {\"total\": ...}: %v", rec.Body.String(), err)
	}
	return out.Total
}

func TestCounterSequence(t *testing.T) {
	h := NewCounter()
	if got := getTotal(t, h); got != 0 {
		t.Fatalf("свежий счётчик: total = %d, ожидалось 0", got)
	}
	if rec := postAdd(t, h, `{"n": 5}`); rec.Code != http.StatusOK {
		t.Fatalf("POST /add {\"n\": 5}: статус = %d, ожидалось 200", rec.Code)
	}
	if rec := postAdd(t, h, `{"n": 7}`); rec.Code != http.StatusOK {
		t.Fatalf("POST /add {\"n\": 7}: статус = %d, ожидалось 200", rec.Code)
	}
	if got := getTotal(t, h); got != 12 {
		t.Errorf("после +5 и +7: total = %d, ожидалось 12", got)
	}
}

func TestCounterIndependent(t *testing.T) {
	first := NewCounter()
	second := NewCounter()
	if rec := postAdd(t, first, `{"n": 100}`); rec.Code != http.StatusOK {
		t.Fatalf("POST /add первому счётчику: статус = %d, ожидалось 200", rec.Code)
	}
	if got := getTotal(t, second); got != 0 {
		t.Errorf("второй счётчик: total = %d, ожидалось 0 — каждый NewCounter независим, глобальные переменные не подойдут", got)
	}
	if got := getTotal(t, first); got != 100 {
		t.Errorf("первый счётчик: total = %d, ожидалось 100", got)
	}
}

func TestCounterBadJSON(t *testing.T) {
	h := NewCounter()
	rec := postAdd(t, h, `{оторвало скобку`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /add с кривым JSON: статус = %d, ожидалось 400", rec.Code)
	}
	if got := getTotal(t, h); got != 0 {
		t.Errorf("после кривого JSON: total = %d, ожидалось 0 — ошибка не должна менять счётчик", got)
	}
}

func TestCounterWrongMethod(t *testing.T) {
	h := NewCounter()
	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /add: статус = %d, ожидалось 405 — метод указывается в шаблоне: \"POST /add\"", rec.Code)
	}
}

func TestCounterConcurrent(t *testing.T) {
	h := NewCounter()
	const workers = 100
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/add", strings.NewReader(`{"n": 1}`))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
		}()
	}
	wg.Wait()
	if got := getTotal(t, h); got != workers {
		t.Errorf("после %d конкурентных POST /add: total = %d, ожидалось %d — защищайте сумму мьютексом", workers, got, workers)
	}
}
