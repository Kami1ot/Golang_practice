Задача со звёздочкой: ждать можно — но не вечно.

Реализуйте функцию:

```go
func Eventually(result <-chan string, timeout time.Duration) (string, error)
```

Она ждёт значение из канала, но не дольше `timeout`:

- значение пришло вовремя — вернуть его и `nil`;
- время вышло — вернуть пустую строку и ошибку истёкшего срока (ту самую, стандартную, — `context.DeadlineExceeded`).

Используйте `context.WithTimeout` — весь смысл задачи в связке «контекст со сроком + select». Не забудьте `defer cancel()`.

Пример работы:

```go
fast := make(chan string, 1)
fast <- "успел!"
Eventually(fast, time.Second) // "успел!", nil

slow := make(chan string) // молчит
_, err := Eventually(slow, 50*time.Millisecond)
errors.Is(err, context.DeadlineExceeded) // true
```

Этот паттерн — щит любого сервиса: ни один медленный ответ не подвесит программу навсегда. Тесты проверяют оба исхода и тип ошибки через `errors.Is`.

Не меняйте имя и сигнатуру функции.
