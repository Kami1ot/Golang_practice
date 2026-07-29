package main

// SumPool суммирует значения из закрытого канала силами workers горутин.
// Общий итог защитите мьютексом или соберите через канал результатов.
func SumPool(jobs <-chan int, workers int) int {
	total := 0
	// TODO: K воркеров с range по jobs + безопасная сборка суммы
	for v := range jobs {
		total += v // пока один поток — переделайте на пул
	}
	return total
}
