package main

// User — профиль пользователя.
// Расставьте json-теги по условию: name, email (omitempty),
// age (omitempty), friends (omitempty); Token в JSON не попадает никогда.
type User struct {
	Name    string
	Email   string
	Age     int
	Friends []string
	Token   string
}

// EncodeUser сериализует профиль в компактную JSON-строку.
func EncodeUser(u User) (string, error) {
	// TODO: json.Marshal + string(b); ошибку пробросить
	return "", nil
}

// DecodeUser разбирает JSON-строку в User.
func DecodeUser(s string) (User, error) {
	// TODO: json.Unmarshal в локальную переменную (указатель!)
	return User{}, nil
}
