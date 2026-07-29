package main

import (
	"fmt"
	"net/http"
)

// HelloHandler отвечает приветствием:
//
//	GET /hello?name=Ада → "Привет, Ада!"
//	GET /hello          → "Привет, гость!"
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: достать name из r.URL.Query(), пустое имя заменить на "гость"
	fmt.Fprint(w, "TODO")
}
