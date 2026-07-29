package main

import "testing"

func TestDepositAndWithdrawChain(t *testing.T) {
	balance := 0
	if !Deposit(&balance, 100) {
		t.Errorf("Deposit(&balance, 100) = false, ожидалось true — пополнение на положительную сумму должно проходить")
	}
	if balance != 100 {
		t.Errorf("после Deposit(&balance, 100) на пустом счёте баланс = %d, ожидалось 100", balance)
	}
	if !Deposit(&balance, 50) {
		t.Errorf("Deposit(&balance, 50) = false, ожидалось true")
	}
	if balance != 150 {
		t.Errorf("после второго пополнения баланс = %d, ожидалось 150 (100 + 50)", balance)
	}
	if !Withdraw(&balance, 30) {
		t.Errorf("Withdraw(&balance, 30) = false, ожидалось true — на счёте 150, снять 30 можно")
	}
	if balance != 120 {
		t.Errorf("после Withdraw(&balance, 30) баланс = %d, ожидалось 120 (150 - 30)", balance)
	}
}

func TestDepositRejectsNonPositive(t *testing.T) {
	balance := 40
	if Deposit(&balance, 0) {
		t.Errorf("Deposit(&balance, 0) = true, ожидалось false — нулевая сумма не принимается")
	}
	if Deposit(&balance, -10) {
		t.Errorf("Deposit(&balance, -10) = true, ожидалось false — отрицательная сумма не принимается")
	}
	if balance != 40 {
		t.Errorf("после отклонённых пополнений баланс = %d, ожидалось 40 — при отказе баланс меняться не должен", balance)
	}
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	balance := 50
	if Withdraw(&balance, 51) {
		t.Errorf("Withdraw(&balance, 51) = true, ожидалось false — на счёте всего 50, столько снять нельзя")
	}
	if balance != 50 {
		t.Errorf("после отклонённого снятия баланс = %d, ожидалось 50 — при отказе баланс не трогаем", balance)
	}
}

func TestWithdrawRejectsNonPositive(t *testing.T) {
	balance := 50
	if Withdraw(&balance, 0) {
		t.Errorf("Withdraw(&balance, 0) = true, ожидалось false — нулевая сумма не принимается")
	}
	if Withdraw(&balance, -5) {
		t.Errorf("Withdraw(&balance, -5) = true, ожидалось false — отрицательная сумма не принимается")
	}
	if balance != 50 {
		t.Errorf("после отклонённых снятий баланс = %d, ожидалось 50 — при отказе баланс меняться не должен", balance)
	}
}

func TestWithdrawToZero(t *testing.T) {
	balance := 70
	if !Withdraw(&balance, 70) {
		t.Errorf("Withdraw(&balance, 70) = false, ожидалось true — снять весь баланс целиком можно (сравнение строгое)")
	}
	if balance != 0 {
		t.Errorf("после снятия всей суммы баланс = %d, ожидалось 0", balance)
	}
}
