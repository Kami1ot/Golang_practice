package main

// Account — счёт с балансом (в целых рублях).
type Account struct {
	Balance int
}

// SavingsAccount — накопительный счёт: встроенный Account
// плюс ставка Rate (процент, целое число: 10 значит 10%).
type SavingsAccount struct {
	Account
	Rate int
}

// Deposit прибавляет amount к балансу. Валидации не нужно.
func (a *Account) Deposit(amount int) {
	// TODO
}

// AddInterest начисляет проценты: увеличивает Balance
// на Balance*Rate/100 (деление целочисленное).
func (s *SavingsAccount) AddInterest() {
	// TODO
}
