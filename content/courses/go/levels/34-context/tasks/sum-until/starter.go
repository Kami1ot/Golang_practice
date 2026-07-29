package main

import "context"

// SumUntil суммирует значения из ch, пока канал не закроют
// или контекст не отменят, — и возвращает накопленное.
func SumUntil(ctx context.Context, ch <-chan int) int {
	total := 0
	// TODO: for + select: ветка приёма с ok-идиомой и ветка ctx.Done()
	return total
}
