Задача со звёздочкой: указатель плюс возвращаемое значение. Смоделируйте банковский счёт: баланс лежит в обычной `int`-переменной, а операции с ним выполняют две функции.

1. `Deposit(balance *int, amount int) bool` — пополнение. При `amount <= 0` операция отклоняется: вернуть `false` и **ничего не менять**. Иначе прибавить `amount` к балансу и вернуть `true`.
2. `Withdraw(balance *int, amount int) bool` — снятие. Отклоняется (`false`, баланс не тронут), если `amount <= 0` **или** денег не хватает (`amount > *balance`). Иначе списать `amount` и вернуть `true`.

```go
balance := 100
Deposit(&balance, 50)   // true,  balance == 150
Withdraw(&balance, 200) // false, balance == 150 — не хватает
Withdraw(&balance, 150) // true,  balance == 0 — снять всё до нуля можно
Deposit(&balance, -5)   // false, balance == 0
```

Требования:

- при отказе баланс остаётся ровно таким, каким был до вызова;
- снять весь баланс целиком (`amount == *balance`) — можно, счёт уходит в ноль;
- не меняйте имена и сигнатуры функций — тесты ищут именно `Deposit(balance *int, amount int) bool` и `Withdraw(balance *int, amount int) bool`.
