Разминка: верните ошибку, которую можно узнать.

В заготовке объявлена сигнальная ошибка (менять её нельзя):

```go
var ErrUserNotFound = errors.New("пользователь не найден")
```

Реализуйте функцию:

```go
func FindAge(users map[string]int, name string) (int, error)
```

Правила:

- имя есть в карте — верните возраст и `nil`;
- имени нет — верните `0` и **именно** `ErrUserNotFound` (то самое значение, не копию!).

Пример работы:

```go
users := map[string]int{"Аня": 25}

FindAge(users, "Аня")                          // 25, nil
_, err := FindAge(users, "Боря")               // 0, ErrUserNotFound
errors.Is(err, ErrUserNotFound)                // true
```

Подвох уровня: если внутри функции написать `errors.New("пользователь не найден")`, текст будет тем же, но `errors.Is` вернёт `false` — каждое `errors.New` создаёт новое значение. Sentinel возвращают, а не пересоздают.

Не меняйте объявление ErrUserNotFound, имя и сигнатуру функции.
