Задача со звёздочкой: спроектируйте ошибки маленького сервиса профилей — все три инструмента уровня разом.

В заготовке даны sentinel-ошибки и тип (менять нельзя):

```go
var (
	ErrEmptyName = errors.New("имя не задано")
	ErrBanned    = errors.New("пользователь заблокирован")
)

type NotFoundError struct {
	Name string
}
```

Реализуйте:

1. Метод `(e NotFoundError) Error() string` — строго `профиль "vasya" не найден`
   (имя в кавычках — используйте глагол `%q`).
2. Функцию `OpenProfile(statuses map[string]string, name string) error`:

| Ситуация | Что вернуть |
|---|---|
| `name` — пустая строка | `ErrEmptyName` |
| имени нет в карте | `NotFoundError{Name: name}` |
| статус `"banned"` | `ErrBanned`, обёрнутый с контекстом: текст `доступ к vasya: пользователь заблокирован` |
| иначе | `nil` |

Пример работы:

```go
statuses := map[string]string{"anna": "active", "vasya": "banned"}

OpenProfile(statuses, "anna")   // nil
OpenProfile(statuses, "")       // ErrEmptyName
OpenProfile(statuses, "kot")    // NotFoundError — errors.As найдёт, Name == "kot"
err := OpenProfile(statuses, "vasya")
errors.Is(err, ErrBanned)       // true — сквозь обёртку
```

Не меняйте объявления ошибок и типа, имя и сигнатуру функции.
