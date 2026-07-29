package main

// QuotaError — ошибка «квота исчерпана»: сколько стало и каков лимит.
// Не меняйте объявление типа.
type QuotaError struct {
	Used, Limit int
}

// Error возвращает сообщение вида "квота исчерпана: 115 из 100".
// Понадобится fmt.Sprintf — добавьте import "fmt".
func (e QuotaError) Error() string {
	// TODO: соберите сообщение из полей
	return ""
}

// Upload проверяет, влезет ли файл размера size при занятых used из limit.
// Перебор — QuotaError с объёмом ПОСЛЕ загрузки; иначе nil.
func Upload(used, limit, size int) error {
	// TODO: сравните used+size с limit
	return nil
}
