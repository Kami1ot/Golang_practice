package main

import "sync"

// Register — потокобезопасная таблица очков. Не меняйте объявление.
type Register struct {
	mu     sync.Mutex
	scores map[string]int
}

// NewRegister возвращает готовый к работе реестр (карта создана!).
func NewRegister() *Register {
	// TODO: не забудьте создать карту
	return &Register{}
}

// Add добавляет игроку очки (безопасно для горутин).
func (r *Register) Add(name string, points int) {
	// TODO: критическая секция вокруг записи в карту
}

// Total возвращает сумму очков игрока; 0 — если не играл.
func (r *Register) Total(name string) int {
	// TODO: чтение — тоже под замком
	return 0
}
