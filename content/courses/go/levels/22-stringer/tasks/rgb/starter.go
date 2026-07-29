package main

// RGB — цвет из трёх компонент 0–255.
// Не меняйте объявление типа.
type RGB struct {
	R, G, B int
}

// String возвращает цвет в формате "rgb(255, 128, 0)".
// Понадобится fmt.Sprintf — добавьте import "fmt".
func (c RGB) String() string {
	// TODO: соберите строку через fmt.Sprintf
	return ""
}
