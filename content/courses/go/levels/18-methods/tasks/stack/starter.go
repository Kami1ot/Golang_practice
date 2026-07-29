package main

// Stack — стек целых чисел (LIFO: последним пришёл — первым ушёл).
// Вершина стека — конец слайса items. Не меняйте объявление типа.
type Stack struct {
	items []int
}

// Push кладёт x на вершину стека.
func (s *Stack) Push(x int) {
	// TODO: допишите x в конец s.items
}

// Pop снимает верхний элемент и возвращает его.
// Для пустого стека возвращает (0, false).
func (s *Stack) Pop() (int, bool) {
	// TODO: проверьте пустоту, снимите последний элемент
	return 0, false
}

// Len возвращает количество элементов в стеке.
func (s Stack) Len() int {
	// TODO: длина s.items
	return 0
}
