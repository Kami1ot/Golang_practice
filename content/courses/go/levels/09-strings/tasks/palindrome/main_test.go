package main

import "testing"

func TestPalindromeCyrillic(t *testing.T) {
	if !IsPalindrome("шалаш") {
		t.Errorf("IsPalindrome(\"шалаш\") = false, ожидалось true — «шалаш» палиндром. Проверьте, что сравниваете руны, а не байты")
	}
}

func TestNotPalindrome(t *testing.T) {
	if IsPalindrome("привет") {
		t.Errorf("IsPalindrome(\"привет\") = true, ожидалось false")
	}
}

func TestPalindromeEmpty(t *testing.T) {
	if !IsPalindrome("") {
		t.Errorf("IsPalindrome(\"\") = false, ожидалось true — пустая строка считается палиндромом")
	}
}

func TestPalindromeSingleRune(t *testing.T) {
	if !IsPalindrome("я") {
		t.Errorf("IsPalindrome(\"я\") = false, ожидалось true — один символ всегда палиндром")
	}
}

func TestPalindromeLatinEven(t *testing.T) {
	if !IsPalindrome("abba") {
		t.Errorf("IsPalindrome(\"abba\") = true ожидалось, получено false")
	}
	if IsPalindrome("abab") {
		t.Errorf("IsPalindrome(\"abab\") = false ожидалось, получено true")
	}
}

func TestPalindromeCaseSensitive(t *testing.T) {
	if IsPalindrome("Аня") {
		t.Errorf("IsPalindrome(\"Аня\") = true, ожидалось false — регистр учитывается, «А» и «я» разные руны")
	}
}

func TestPalindromeCyrillicNotEqualBytes(t *testing.T) {
	if !IsPalindrome("топот") {
		t.Errorf("IsPalindrome(\"топот\") = false, ожидалось true — «топот» палиндром")
	}
}
