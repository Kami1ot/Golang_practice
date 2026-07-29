package main

// Person — человек: имя и возраст. Объявление типа не менять.
type Person struct {
	Name string
	Age  int
}

// Oldest возвращает имя самого старшего человека в слайсе.
// При равных возрастах — имя первого из равных.
// Слайс people гарантированно непустой.
func Oldest(people []Person) string {
	// TODO: переберите слайс, запоминая самого старшего
	return ""
}
