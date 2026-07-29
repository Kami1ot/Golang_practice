package main

// User — данные нового пользователя для отправки на сервер.
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUser отправляет u POST-запросом (JSON) на baseURL+"/users".
// Сервер отвечает 201 и телом {"id":N} — вернуть N.
// Любой другой статус — 0 и ошибка с кодом статуса в тексте.
func CreateUser(baseURL string, u User) (int, error) {
	// TODO: json.Marshal → http.Post → статус 201 → id из JSON-ответа
	return 0, nil
}
