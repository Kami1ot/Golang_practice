# Встраивание и композиция — собираем типы из кусочков

Структуру в структуру вы уже вкладывали: поле с именем, доступ цепочкой `car.Motor.Power`. Работает, но многословно. В Go есть приём получше — положить тип внутрь типа **без имени поля**. Это встраивание (*embedding*): внешний тип получает поля и методы внутреннего «как свои». Заодно это ответ Go на вечный вопрос про наследование: наследования в Go **нет**, вместо него — композиция (*composition*). Что это значит на практике — разберём до винтика.

## Шаг назад: поле-структура с именем

Так вы собирали типы раньше — у поля есть имя:

```go
type Engine struct{ Power int }

type Car struct {
	Motor Engine // обычное поле: имя + тип
	Brand string
}

car := Car{Motor: Engine{Power: 150}, Brand: "GoMobile"}
fmt.Println(car.Motor.Power) // 150
```

Каждое обращение — через имя поля: `car.Motor.Power`. Три звена ради одного числа, а если у `Engine` есть методы — тоже только `car.Motor.Start()`.

## Встраивание: поле без имени

Уберите имя поля и оставьте только тип — это и есть встраивание (*embedding*):

```go
type Car struct {
	Engine       // встроенный тип: имени нет, только тип
	Brand string
}
```

Поле никуда не исчезло — его имя теперь совпадает с именем типа: `Engine`. Но у такой записи появляется суперсила.

## Продвижение полей

Поля встроенного типа **продвигаются** (*promotion*) во внешний: к ним можно обращаться напрямую, как будто они объявлены прямо в `Car`. Полное имя при этом тоже работает — оно нужно, когда работаете со встроенной структурой целиком:

```go
package main

import "fmt"

type Engine struct{ Power int }

type Car struct {
	Engine
	Brand string
}

func main() {
	car := Car{Engine: Engine{Power: 150}, Brand: "GoMobile"}
	fmt.Println(car.Power)        // 150 — коротко: поле продвинулось
	fmt.Println(car.Engine.Power) // 150 — полное имя тоже работает

	car.Engine = Engine{Power: 200} // заменить весь двигатель разом
	fmt.Println(car.Engine)         // {200} — напечатать только его
}
```

`car.Power` и `car.Engine.Power` — это **одно и то же поле**, просто короткая запись удобнее.

> [!WARNING]
> В композитном литерале продвижение не действует: `Car{Power: 150}` — ошибка компиляции `unknown field Power in struct literal`. Инициализируют встроенное поле по имени типа: `Car{Engine: Engine{Power: 150}, Brand: "GoMobile"}`. Продвижение работает при чтении и записи через точку, но не при создании.

## Продвижение методов

Продвигаются не только поля — **методы** встроенного типа тоже вызываются напрямую у внешнего. Для `Car` мы не напишем ни строчки, а метод `Start` у него появится — именно так в Go «переиспользуют поведение»:

```go
package main

import "fmt"

type Engine struct{ Power int }

func (e Engine) Start() string {
	return fmt.Sprintf("двигатель %d л.с. заведён", e.Power)
}

type Car struct {
	Engine
	Brand string
}

func main() {
	car := Car{Engine: Engine{Power: 90}, Brand: "Gopher"}
	fmt.Println(car.Start())        // двигатель 90 л.с. заведён
	fmt.Println(car.Engine.Start()) // то же самое через полное имя
}
```

## Встроить можно несколько типов

Внешний тип собирает поля и методы **всех** встроенных:

```go
package main

import "fmt"

type Engine struct{ Power int }
type Radio struct{ Station string }

type Car struct {
	Engine
	Radio
	Brand string
}

func main() {
	car := Car{Engine: Engine{Power: 110}, Radio: Radio{Station: "Go FM"}, Brand: "GoMobile"}
	fmt.Println(car.Power)   // 110 — продвинулось из Engine
	fmt.Println(car.Station) // Go FM — продвинулось из Radio
}
```

## Затенение: своё имя сильнее продвинутого

Если у внешнего типа есть **собственное** поле или метод с тем же именем, что у встроенного, — своё побеждает. Это затенение (*shadowing*). Внутреннее при этом не пропадает — оно доступно через полное имя:

```go
package main

import "fmt"

type Person struct{ Name string }

type Employee struct {
	Person
	Name string // собственное поле затеняет продвинутое
}

func main() {
	e := Employee{Person: Person{Name: "Ира"}, Name: "irina_dev"}
	fmt.Println(e.Name)        // irina_dev — своё поле победило
	fmt.Println(e.Person.Name) // Ира — внутреннее живо, доступно по полному имени
}
```

С методами то же самое — собственный метод внешнего типа перекрывает продвинутый:

```go
package main

import "fmt"

type Animal struct{ Name string }

func (a Animal) Voice() string {
	return a.Name + " молчит"
}

type Cat struct{ Animal }

// Собственный Voice затеняет продвинутый Animal.Voice.
func (c Cat) Voice() string {
	return c.Name + " мяукает"
}

func main() {
	c := Cat{Animal: Animal{Name: "Барсик"}}
	fmt.Println(c.Voice())        // Барсик мяукает — сработал метод Cat
	fmt.Println(c.Animal.Voice()) // Барсик молчит — внутренний вызывается полным именем
}
```

## Конфликт: два одинаковых продвинутых имени

А если **два** встроенных типа принесли поле с одним именем и своего у внешнего нет? Продвижение ломается — но не сразу:

```go
package main

import "fmt"

type A struct{ Name string }
type B struct{ Name string }

type AB struct {
	A
	B
}

func main() {
	x := AB{A: A{Name: "первый"}, B: B{Name: "второй"}}
	fmt.Println(x.A.Name) // первый — полное имя работает
	fmt.Println(x.B.Name) // второй
	// fmt.Println(x.Name) // ошибка компиляции: ambiguous selector x.Name
}
```

> [!NOTE]
> Конфликт имён — ошибка компиляции **только при обращении к короткому имени**. Сам тип `AB` объявлять можно, программа компилируется и работает, пока вы пользуетесь полными именами `x.A.Name` и `x.B.Name`. Компилятор не гадает, «какое из двух» вы имели в виду, — он просто требует уточнить.

## Продвигаются и методы с pointer receiver

Методы с указательным получателем (*pointer receiver*) с прошлого уровня тоже продвигаются — и честно меняют встроенную часть:

```go
package main

import "fmt"

type Counter struct{ Total int }

func (c *Counter) Inc() { // указательный получатель: метод меняет счётчик
	c.Total++
}

type Clicker struct {
	Counter
	Label string
}

func main() {
	btn := Clicker{Label: "кнопка"}
	btn.Inc() // метод *Counter продвинулся: Go сам возьмёт адрес btn.Counter
	btn.Inc()
	fmt.Println(btn.Total) // 2
}
```

Как и на прошлом уровне, `btn` должна быть переменной (адресуемым значением) — тогда Go сам подставит `(&btn.Counter).Inc()`.

## Это НЕ наследование

Встраивание внешне похоже на наследование из других языков, но работает иначе — и это надо прочувствовать:

```go
package main

import "fmt"

type Base struct{ Name string }

func (b Base) Who() string {
	return "я " + b.Name
}

func (b Base) Intro() string {
	return "Это " + b.Who() // зовёт Who типа Base — и только его
}

type Admin struct {
	Base
	Level int
}

// Собственный Who затеняет продвинутый…
func (a Admin) Who() string {
	return "админ " + a.Name
}

func main() {
	a := Admin{Base: Base{Name: "Кама"}, Level: 9}
	fmt.Println(a.Who())   // админ Кама — короткое имя берёт метод Admin
	fmt.Println(a.Intro()) // Это я Кама — Intro «переопределения» не заметил!
}
```

> [!WARNING]
> Встроенный тип **ничего не знает о внешнем**. Метод `Base.Intro` живёт внутри `Base`: он видит только поля `Base` (до `a.Level` ему не дотянуться — будет `undefined`) и зовёт только методы `Base`. В языке с наследованием `Intro` вызвал бы «переопределённый» `Who` — в Go виртуальных методов нет, затенение действует только снаружи, при вызове через внешний тип.

Именно поэтому про встраивание не говорят «наследование». Правильные слова: `Car` **содержит** `Engine`, а не «является» им.

## Композиция — путь Go

Итог философии: вместо иерархий-деревьев из классов Go предлагает собирать типы из маленьких независимых кусочков, как из конструктора. `Counter` умеет считать, `Radio` — играть, `Engine` — ехать; встроили нужные — получили тип с готовым поведением, ничего не переписывая. Это и есть композиция (*composition*): «собери из частей» вместо «отними у предка».

## Шпаргалка: поле с именем vs встраивание

| | Поле с именем | Встраивание |
|---|---|---|
| Объявление | `Motor Engine` | `Engine` |
| Чтение поля | `car.Motor.Power` | `car.Power` (и `car.Engine.Power`) |
| Вызов метода | `car.Motor.Start()` | `car.Start()` (и `car.Engine.Start()`) |
| В литерале | `Motor: Engine{…}` | `Engine: Engine{…}` — имя = имя типа |
| Когда выбирать | деталь среди прочих: доступ по имени читается яснее | внутренний тип — ядро внешнего: его поля и методы нужны напрямую |

## Типичные ошибки новичка

| Код | Что случится | Как чинить |
|---|---|---|
| `Car{Power: 150}` | `unknown field Power in struct literal` | в литерале встроенное поле зовут по имени типа: `Car{Engine: Engine{Power: 150}}` |
| `x.Name` при двух встроенных с полем `Name` | `ambiguous selector x.Name` | уточнить полным именем: `x.A.Name` или `x.B.Name` |
| метод встроенного типа читает поле внешнего | `undefined` — внутренний тип о внешнем не знает | перенести метод на внешний тип или передать значение параметром |
| надежда, что `Intro` из примера вызовет затеняющий `Who` | внутренние методы зовут только внутренние | это не наследование: общее поведение собирайте методами внешнего типа |

## Запомнить

- Встраивание — поле без имени: `type Car struct { Engine; Brand string }`; имя поля = имя типа.
- Поля и методы встроенного типа продвигаются: `car.Power`, `car.Start()`. Полное имя `car.Engine.Power` всегда доступно и нужно для работы с частью целиком.
- В композитном литерале продвижения нет: `Car{Engine: Engine{Power: 150}}`.
- Затенение: своё поле/метод внешнего типа сильнее продвинутого; внутреннее — через полное имя.
- Конфликт двух одинаковых продвинутых имён — ошибка только при обращении к короткому имени.
- Методы с pointer receiver тоже продвигаются: вызов через внешнюю переменную меняет встроенную часть.
- Это не наследование: встроенный тип не знает о внешнем, его методы видят только свои поля. Путь Go — композиция.
