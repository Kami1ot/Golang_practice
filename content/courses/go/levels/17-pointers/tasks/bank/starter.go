package main

// Deposit пытается пополнить счёт на amount.
// При amount <= 0 возвращает false и не меняет баланс.
// Иначе прибавляет amount к балансу и возвращает true.
func Deposit(balance *int, amount int) bool {
	// TODO: проверка суммы, изменение баланса по указателю
	return false
}

// Withdraw пытается снять amount со счёта.
// При amount <= 0 или нехватке денег (amount > *balance)
// возвращает false и не меняет баланс.
// Иначе списывает amount и возвращает true.
func Withdraw(balance *int, amount int) bool {
	// TODO: две проверки, затем списание по указателю
	return false
}
