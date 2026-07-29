package main

// User — пользователь из JSON-API.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FetchUser делает GET baseURL+"/users/<id>" и разбирает JSON-ответ.
// Статус не 200 — нулевой User и ошибка с кодом статуса в тексте.
func FetchUser(baseURL string, id int) (User, error) {
	// TODO: собрать URL → http.Get → статус → json.NewDecoder(resp.Body).Decode
	return User{}, nil
}
