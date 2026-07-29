package main

// Person — человек: имя и возраст.
type Person struct {
	Name string
	Age  int
}

// Employee — сотрудник: встроенный Person плюс компания.
type Employee struct {
	Person
	Company string
}

// Greet возвращает строку вида "Привет, я Ира".
func (p Person) Greet() string {
	// TODO
	return ""
}

// Card возвращает строку вида "Ира, 30 лет, работает в Гофер-софт".
// Имя и возраст доступны как e.Name и e.Age — поля продвинулись из Person.
func (e Employee) Card() string {
	// TODO
	return ""
}

// Greet возвращает строку вида "Привет, я Ира из Гофер-софт".
// Собственный метод Employee — перекрывает продвинутый Person.Greet.
func (e Employee) Greet() string {
	// TODO
	return ""
}
