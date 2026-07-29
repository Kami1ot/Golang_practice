Основная задача: научите структуру `User` правильно ездить в JSON и обратно.

В заготовке структура объявлена **без тегов** — расставить их часть задачи:

- `Name` → в JSON зовётся `name`, присутствует всегда;
- `Email` → `email`, пустое значение пропускается (*omitempty*);
- `Age` → `age`, нулевой возраст пропускается;
- `Friends` → `friends`, пустой список пропускается;
- `Token` → в JSON **не попадает никогда** (и при разборе игнорируется).

Затем реализуйте две функции:

```go
func EncodeUser(u User) (string, error)
func DecodeUser(s string) (User, error)
```

`EncodeUser` возвращает компактный JSON (обычный `Marshal`, без отступов) строкой; `DecodeUser` разбирает JSON-строку в `User`. Обе честно пробрасывают ошибку пакета `json`, если что-то пошло не так.

Пример работы:

```go
u := User{Name: "gopher", Email: "go@example.com", Age: 3,
	Friends: []string{"rob", "ken"}, Token: "секрет"}

EncodeUser(u)
// {"name":"gopher","email":"go@example.com","age":3,"friends":["rob","ken"]}
// Token исчез: json:"-"

EncodeUser(User{Name: "нео"})
// {"name":"нео"} — пустые email/age/friends спрятал omitempty

DecodeUser(`{"name":"ann","age":19}`)
// User{Name: "ann", Age: 19} — остальные поля остались нулевыми
```

Тесты сверяют JSON **посимвольно** (порядок полей = порядку объявления — не переставляйте их), проверяют, что `Token` не утекает наружу и не заполняется при разборе, а кривой JSON возвращает ошибку.

Не меняйте имена и сигнатуры функций, имена и порядок полей структуры.
