package main

import "testing"

func TestAccountDeposit(t *testing.T) {
	a := Account{}
	a.Deposit(100)
	a.Deposit(50)
	if a.Balance != 150 {
		t.Errorf("после Deposit(100) и Deposit(50) баланс Account = %d, ожидалось 150 — Deposit должен прибавлять amount к Balance", a.Balance)
	}
}

func TestSavingsDepositPromoted(t *testing.T) {
	s := SavingsAccount{Rate: 10}
	s.Deposit(200)
	if s.Balance != 200 {
		t.Errorf("после s.Deposit(200) баланс SavingsAccount = %d, ожидалось 200 — метод Deposit с pointer receiver должен продвигаться из Account и менять встроенный счёт", s.Balance)
	}
}

func TestAddInterest(t *testing.T) {
	s := SavingsAccount{Rate: 10}
	s.Deposit(200)
	s.AddInterest()
	if s.Balance != 220 {
		t.Errorf("Deposit(200), затем AddInterest() при Rate=10: баланс = %d, ожидалось 220 (200 + 200*10/100)", s.Balance)
	}
}

func TestAddInterestTruncates(t *testing.T) {
	s := SavingsAccount{Account: Account{Balance: 105}, Rate: 10}
	s.AddInterest()
	if s.Balance != 115 {
		t.Errorf("AddInterest() при Balance=105, Rate=10: баланс = %d, ожидалось 115 — деление целочисленное: 105*10/100 = 10, дробная часть отбрасывается", s.Balance)
	}
}

func TestAddInterestSeveralTimes(t *testing.T) {
	s := SavingsAccount{Account: Account{Balance: 100}, Rate: 50}
	s.AddInterest()
	if s.Balance != 150 {
		t.Errorf("первое AddInterest() при Balance=100, Rate=50: баланс = %d, ожидалось 150", s.Balance)
	}
	s.AddInterest()
	if s.Balance != 225 {
		t.Errorf("второе AddInterest() подряд: баланс = %d, ожидалось 225 — проценты считаются уже от нового баланса 150", s.Balance)
	}
}

func TestZeroRate(t *testing.T) {
	s := SavingsAccount{Account: Account{Balance: 500}, Rate: 0}
	s.AddInterest()
	if s.Balance != 500 {
		t.Errorf("AddInterest() при Rate=0: баланс = %d, ожидалось 500 — нулевая ставка ничего не начисляет", s.Balance)
	}
}
