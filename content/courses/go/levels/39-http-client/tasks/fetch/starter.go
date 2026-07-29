package main

// Fetch скачивает url GET-запросом и возвращает тело ответа строкой.
// Статус не 200 — пустая строка и ошибка с кодом статуса в тексте.
func Fetch(url string) (string, error) {
	// TODO: http.Get → проверить err → defer Body.Close → статус → io.ReadAll
	return "", nil
}
