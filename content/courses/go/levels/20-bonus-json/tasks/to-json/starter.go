package main

import "fmt"

type Player struct {
	Name  string // TODO: добавьте тег, чтобы в JSON поле звалось name
	Score int    // TODO: добавьте тег, чтобы в JSON поле звалось score
}

func main() {
	var p Player
	fmt.Scan(&p.Name, &p.Score)
	// TODO: сериализуйте p через json.Marshal
	// и напечатайте результат как строку
}
