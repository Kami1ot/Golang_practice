package main

// Engine — двигатель: мощность в лошадиных силах.
type Engine struct {
	Power int
}

// Car — машина: встроенный двигатель и марка.
// Engine записан без имени поля — это встраивание.
type Car struct {
	Engine
	Brand string
}

// Describe возвращает строку вида "двигатель 150 л.с.".
func (e Engine) Describe() string {
	// TODO: соберите строку (пригодится fmt.Sprintf — не забудьте import "fmt")
	return ""
}

// FullName возвращает строку вида "Gopher, двигатель 90 л.с.":
// марка, запятая с пробелом, описание двигателя.
func (c Car) FullName() string {
	// TODO: мощность доступна как c.Power, а метод Describe продвинулся из Engine
	return ""
}
