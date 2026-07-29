# План курса «Язык Go» — дорожная карта контента

Структура: ОДИН курс `go` («Язык Go»), уровни сгруппированы в секции на карте
(поле `group` в level.json). Отдельные курсы заводим только под другие большие
темы (другой язык, алгоритмы, инструменты).

Отмечаем заведение контента: `[x]` — уровень готов и проверен в браузере.
Детали формата — AUTHORING.md.

## Секция «Основы» ✅
- [x] 01-hello — Привет, Go! (программа, go run/build, Println, комментарии, gofmt)
- [x] 02-variables — Переменные и типы (var, :=, типы, конверсии, константы, Scan)
- [x] 03-conditions — Условия (bool, if/else, if с объявлением, switch, %)
- [x] 04-bonus-fmt — Бонус: форматированный вывод (Printf, глаголы, Sprintf)

## Секция «Циклы и коллекции» ✅
- [x] 05-for — Цикл for (три формы, счётчики, аккумуляторы)
- [x] 06-range — range, break и continue (включая range по int из Go 1.22+)
- [x] 07-slices — Слайсы (make, append, len/cap, срезы, копирование)
- [x] 08-maps — Карты (ok-идиома, delete, range, счётчик частот)
- [x] 09-strings — Строки и руны (UTF-8, []rune, пакет strings, strconv)
- [x] 10-bonus-nested — Бонус: вложенные циклы (фигуры, метки, O(n²))

## Секция «Функции» ✅
- [x] 11-func-basics — Сигнатуры, параметры, множественный возврат
- [x] 12-func-named-variadic — Именованные результаты, variadic (...)
- [x] 13-closures — Функции как значения, замыкания (closures)
- [x] 14-defer — defer: отложенное выполнение, LIFO
- [x] 15-bonus-recursion — Бонус: рекурсия

## Секция «Структуры и указатели» ✅
- [x] 16-structs — Структуры (struct): объявление, литералы, доступ
- [x] 17-pointers — Указатели (pointer): &, *, зачем нужны
- [x] 18-methods — Методы: value vs pointer receiver
- [x] 19-embedding — Встраивание (embedding), композиция
- [x] 20-bonus-json — Бонус: JSON-теги, encoding/json

## Секция «Интерфейсы» ✅
- [x] 21-interfaces — Интерфейс (interface): неявная реализация
- [x] 22-stringer — fmt.Stringer, error как интерфейс
- [x] 23-readers-writers — io.Reader / io.Writer
- [x] 24-type-switch — Пустой интерфейс, type assertion, type switch
- [x] 25-bonus-sort — Бонус: sort.Slice и сортировка

## Секция «Обработка ошибок» ✅
- [x] 26-errors — error, errors.New, fmt.Errorf, идиома if err != nil
- [x] 27-wrapping — Оборачивание: %w, errors.Is / errors.As
- [x] 28-custom-errors — Собственные типы ошибок, sentinel errors
- [x] 29-panic — panic/recover: когда можно, когда нельзя

## Секция «Горутины и каналы» ✅
- [x] 30-goroutines — Горутины (goroutine), go-ключевое слово
- [x] 31-channels — Каналы (channel): отправка, приём, блокировки
- [x] 32-buffered-select — Буферизованные каналы, select
- [x] 33-sync — sync.WaitGroup, sync.Mutex, гонки данных
- [x] 34-context — context: отмена и таймауты
- [x] 35-bonus-workerpool — Бонус: пул воркеров (worker pool)

## Секция «Стандартная библиотека» ✅
- [x] 36-time — time: моменты, длительности, форматирование
- [x] 37-files — Файлы: os, bufio, чтение/запись
- [x] 38-json — encoding/json: Marshal/Unmarshal, теги (углубление после 20-bonus-json)
- [x] 39-http-client — net/http клиент: GET/POST, разбор ответов (задачи на httptest)
- [x] 40-http-server — net/http сервер: handlers, ServeMux (паттерны Go 1.22+)

## Секция «Дженерики»
- [ ] 41-type-params — Параметры типов, ограничения (constraints)
- [ ] 42-generic-practice — Практика: Map/Filter/Reduce, обобщённые структуры

## Секция «Тестирование»
- [ ] 43-testing — go test, t.Errorf, соглашения
- [ ] 44-table-tests — Табличные тесты, подтесты t.Run
- [ ] 45-bonus-bench — Бонус: бенчмарки

## Секция «Капстоны»
- [ ] 46-capstone-cli — CLI-утилита: todo-список с файловым хранением
- [ ] 47-capstone-api — JSON-API сервер с net/http
- [ ] 48-capstone-parser — Конкурентный парсер логов

## Будущие отдельные курсы (другие темы)
- [ ] algorithms — Алгоритмы и структуры данных на Go
- [ ] (по запросу пользователя: другие языки, SQL, инструменты…)
