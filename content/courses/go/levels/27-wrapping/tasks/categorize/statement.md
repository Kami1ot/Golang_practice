Основная задача: на вход прилетают ошибки из разных глубин системы — уже обёрнутые, по несколько слоёв. Разложите их по полочкам.

В заготовке даны (менять не нужно):

```go
var ErrNotFound = errors.New("не найдено")

type ValidationError struct {
	Field string
}
// метод Error() уже реализован
```

Реализуйте функцию:

```go
func Categorize(err error) string
```

Правила категоризации:

| Вход | Результат |
|---|---|
| `nil` | `ок` |
| в цепочке есть `ErrNotFound` | `не найдено` |
| в цепочке есть `ValidationError` | `невалидное поле: <Field>` |
| всё остальное | `неизвестная ошибка` |

Пример работы:

```go
Categorize(nil)                                              // "ок"
Categorize(fmt.Errorf("запрос: %w", ErrNotFound))            // "не найдено"
Categorize(fmt.Errorf("вход: %w", ValidationError{"email"})) // "невалидное поле: email"
Categorize(errors.New("что-то ещё"))                         // "неизвестная ошибка"
```

Ошибки приходят обёрнутыми — иногда в два-три слоя, — поэтому наивные `==` и type assertion не сработают: нужны `errors.Is` и `errors.As`.

Не меняйте объявления ErrNotFound и ValidationError, имя и сигнатуру Categorize.
