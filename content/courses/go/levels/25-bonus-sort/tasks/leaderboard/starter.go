package main

// Player — участник рейтинга: имя и очки.
// Не меняйте объявление типа.
type Player struct {
	Name  string
	Score int
}

// Rank сортирует игроков на месте: очки по убыванию,
// при равных очках — имена по алфавиту.
func Rank(players []Player) {
	// TODO: sort.Slice с less-функцией на два ключа (import "sort")
	_ = players
}
