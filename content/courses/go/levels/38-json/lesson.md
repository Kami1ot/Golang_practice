# encoding/json в деле

В прошлом уровне вы научились открывать и читать файлы. А что внутри этих файлов? Удивительно часто — JSON: конфиги сервисов, выгрузки данных, ответы веб-API, журналы событий. Пакет `encoding/json` вы, возможно, уже трогали в бонусном уровне про JSON — сегодня разберёмся, как с ним работают во «взрослых» программах: необязательные поля, вложенные документы, потоковое чтение прямо из файла и разбор JSON, форму которого вы заранее не знаете.

## База за одну минуту

Если проходили бонусный уровень — это повторение; если нет — вот весь минимум, его хватит. JSON-объект `{}` соответствует структуре, массив `[]` — слайсу. Сериализация (*marshalling*) — `json.Marshal`, разбор (*unmarshalling*) — `json.Unmarshal`:

```go
type Movie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

b, err := json.Marshal(Movie{Title: "Хакеры", Year: 1995})
fmt.Println(string(b), err)

var m Movie
err = json.Unmarshal([]byte(`{"title":"Матрица","year":1999}`), &m)
fmt.Println(m.Title, m.Year)
```

```
{"title":"Хакеры","year":1995} <nil>
Матрица 1999
```

Четыре правила, на которых всё держится:

- `Marshal` возвращает `[]byte` и ошибку; печатают через `string(b)`. Поля идут в JSON **в порядке объявления** структуры.
- `Unmarshal` принимает **указатель** — `&m`. Передадите значение без `&` — получите ошибку, а в структуру ничего не запишется.
- Пакет `json` видит только **экспортированные** (*exported*) поля — с большой буквы. Поле `year int` с маленькой пропадёт из JSON молча.
- Имя для JSON задаёт тег структуры (*struct tag*): `` `json:"title"` `` — обратные кавычки, точность до символа; опечатку в теге компилятор не ловит.

И характер `Unmarshal`: он покладистый. Лишнее поле в JSON — молча игнорирует; поля нет в JSON — оставляет в структуре то, что там было. Это «оставляет как было» нам ещё пригодится.

## Опции тегов: omitempty и «-»

В теге после имени можно через запятую дописать опции. Самая ходовая — `omitempty`: «пропусти поле, если значение пустое»:

```go
type Profile struct {
	Name string   `json:"name"`
	Bio  string   `json:"bio,omitempty"`
	Age  int      `json:"age,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

b, _ := json.Marshal(Profile{Name: "gopher", Bio: "пишу на Go", Age: 3, Tags: []string{"go"}})
fmt.Println(string(b))
b, _ = json.Marshal(Profile{Name: "gopher"})
fmt.Println(string(b))
```

```
{"name":"gopher","bio":"пишу на Go","age":3,"tags":["go"]}
{"name":"gopher"}
```

Пустой профиль сжался до одного поля. «Пустым» (*empty*) считается: `0`, `""`, `false`, `nil` (указатель, интерфейс) и слайс или карта **нулевой длины**. Ровно нулевые значения типов — плюс пустые коллекции.

> [!WARNING]
> Вложенная структура со всеми нулевыми полями «пустой» НЕ считается: `omitempty` её не спрячет, в JSON уедет `{"address":{}}`-подобный огрызок. Хотите прятать структуру целиком — сделайте поле указателем (`*Address`): `nil`-указатель уже пустой.

Вторая опция — на самом деле специальное имя — минус: `` `json:"-"` `` означает «этого поля для JSON не существует». Ни в `Marshal`, ни в `Unmarshal` — в обе стороны:

```go
type Account struct {
	Login    string `json:"login"`
	Password string `json:"-"` // секреты не сериализуем НИКОГДА
}

b, _ := json.Marshal(Account{Login: "gopher", Password: "тс-с-с"})
fmt.Println(string(b)) // {"login":"gopher"}
```

Сводка вариантов тега:

| Тег | Что означает |
|---|---|
| `json:"name"` | в JSON поле зовётся `name` |
| `json:"name,omitempty"` | имя `name` + пропуск пустого значения |
| `json:",omitempty"` | имя не задано → берётся Go-имя поля (с большой буквы!), пустое пропускается |
| `json:"omitempty"` | ловушка: это ИМЯ — поле уедет в JSON как `"omitempty"` |
| `json:"-"` | поля для JSON не существует (в обе стороны); имя буквально `-` — экзотика `json:"-,"` |

> [!WARNING]
> Запятая решает всё. `` `json:",omitempty"` `` — опция без имени: поле останется под Go-именем `Score`, хотя вы, скорее всего, хотели `score`. А `` `json:"omitempty"` `` без запятой — имя: в JSON появится поле `"omitempty"`. Обе ошибки компилируются молча — сверяйте теги посимвольно.

## Вложенность: структуры, слайсы, карты

Реальный JSON редко бывает плоским. Хорошая новость: `encoding/json` собирает вложенность любой глубины сам — опишите структуру, остальное не ваша забота:

```go
type Author struct {
	Name string `json:"name"`
}

type Book struct {
	Title   string         `json:"title"`
	Author  Author         `json:"author"`
	Genres  []string       `json:"genres"`
	Ratings map[string]int `json:"ratings"`
}

b := Book{Title: "Go для всех", Author: Author{Name: "Роб"},
	Genres: []string{"учебник", "хит"}, Ratings: map[string]int{"habr": 5, "ozon": 4}}
out, _ := json.Marshal(b)
fmt.Println(string(out))
```

```
{"title":"Go для всех","author":{"name":"Роб"},"genres":["учебник","хит"],"ratings":{"habr":5,"ozon":4}}
```

Структура в структуре стала объектом в объекте, слайс — массивом, а `map[string]int` — JSON-объектом с парами «ключ: число». Обратный путь — тот же `Unmarshal` в `Book`: всё дерево заполнится за один вызов.

> [!NOTE]
> Ключи карты `Marshal` выводит **по алфавиту** (чтобы результат был детерминированным), а поля структуры — в порядке объявления. Если в тестах сверяется точный JSON — помните об этом.

## Указатели-поля: «нет» и «ноль» — разные вещи

Пусть клиент присылает обновление настроек: `{"volume": 0}` — «сделай беззвучно», а `{}` — «громкость не трогай». В поле `Volume int` оба варианта дадут `0` — намерения неразличимы. Решение — поле-указатель:

```go
type Update struct {
	Volume *int `json:"volume"`
}

var a, b, c Update
json.Unmarshal([]byte(`{}`), &a)              // поля нет
json.Unmarshal([]byte(`{"volume":null}`), &b) // явный null
json.Unmarshal([]byte(`{"volume":0}`), &c)    // настоящий ноль

fmt.Println(a.Volume == nil, b.Volume == nil, c.Volume == nil)
fmt.Println(*c.Volume)
```

```
true true false
0
```

Отсутствующее поле и `null` дают `nil`-указатель (между собой они неразличимы — и обычно это ок), а вот пришедший ноль — это уже **не nil**: указатель на честный `0`. Проверили на `nil` — разыменовали — применили. В обратную сторону: `nil`-указатель сериализуется как `null`, а с `omitempty` — исчезает совсем.

Тот же принцип «отсутствующее не трогаем» даёт бесплатные значения по умолчанию — заполните структуру ДО разбора: что пришло в JSON, то перепишется, остальное останется как было:

```go
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

cfg := Config{Host: "localhost", Port: 8080} // умолчания
json.Unmarshal([]byte(`{"port":9000}`), &cfg)
fmt.Println(cfg.Host, cfg.Port)
```

```
localhost 9000
```

## Decoder и Encoder: JSON из потока

До сих пор мы кормили `Unmarshal` срезом байтов. Но JSON чаще живёт в потоке: файл, сетевое соединение, стандартный ввод. Для потоков есть пара `json.NewDecoder` / `json.NewEncoder` — они работают с любыми `io.Reader` и `io.Writer`. А `os.File` из прошлого уровня — как раз `io.Reader`:

```go
f, err := os.Open("config.json")
if err != nil {
	fmt.Println("ошибка:", err)
	return
}
defer f.Close()

var cfg Config
if err := json.NewDecoder(f).Decode(&cfg); err != nil {
	fmt.Println("разбор config.json:", err)
	return
}
```

Ни `ReadAll`, ни промежуточного `[]byte`: декодер читает из файла ровно столько, сколько нужно для одного значения, — большой файл не окажется в памяти целиком. Второй плюс — универсальность: тот же код работает с сетевым соединением, с `os.Stdin`, со `strings.NewReader` в тестах.

`Encoder` — зеркальный брат для записи:

```go
enc := json.NewEncoder(os.Stdout)
enc.Encode(Movie{Title: "Хакеры", Year: 1995})
enc.Encode(Movie{Title: "Матрица", Year: 1999})
```

```
{"title":"Хакеры","year":1995}
{"title":"Матрица","year":1999}
```

> [!NOTE]
> `Encode` сам дописывает `\n` после каждого значения. Два вызова — две строки, по JSON-объекту на каждой. Это не случайность, а формат.

## NDJSON: по объекту на строку

Формат «JSON-объект на строку» зовётся NDJSON (*newline-delimited JSON*) — так пишут журналы событий и выгрузки: файл можно дописывать и читать построчно, не разбирая целиком. Читают его тем же декодером в цикле — метод `More` сообщает, есть ли в потоке следующее значение:

```go
data := `{"title":"Хакеры","year":1995}
{"title":"Матрица","year":1999}`

dec := json.NewDecoder(strings.NewReader(data))
for dec.More() {
	var m Movie
	if err := dec.Decode(&m); err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	fmt.Println(m.Year, m.Title)
}
```

```
1995 Хакеры
1999 Матрица
```

Один декодер — много `Decode`: каждый вызов съедает следующее значение из потока. Строго говоря, декодеру даже не нужны переносы строк — он ест значения подряд через любые пробелы, — но конвенция NDJSON делает файл читаемым и для глаз, и для `grep`.

## JSON неизвестной формы

Иногда структуру не опишешь заранее: API возвращает «какой-то JSON», и надо посмотреть, что внутри. Тогда разбирают в `map[string]any` — а дальше в дело идёт switch по типу (*type switch*) с уровня про `any`:

```go
data := []byte(`{"id": 7, "name": "gopher", "admin": true, "tags": ["go", "json"]}`)

var v map[string]any
json.Unmarshal(data, &v) // err проверяем как всегда — здесь опущено для краткости

fmt.Println(v["name"], v["admin"])
id := v["id"].(float64) // числа приходят ТОЛЬКО как float64
fmt.Println(int(id))

switch t := v["tags"].(type) {
case []any:
	fmt.Println("список из", len(t), "элементов")
case string:
	fmt.Println("строка:", t)
}
```

```
gopher true
7
список из 2 элементов
```

В `any` из JSON попадает всего шесть типов: `string`, `float64`, `bool`, `[]any`, `map[string]any` и `nil`. Заметьте: **все числа — `float64`**, даже «целые с виду»: в JSON нет отдельного целого типа, и декодер не гадает. Написать `v["id"].(int)` — поймать панику (assertion проверяет точный тип, он не конвертирует); правильно — достать `float64` и привести: `int(v["id"].(float64))`.

## json.RawMessage — отложенный разбор ⭐

Гибрид двух миров: часть JSON разобрать сейчас, часть — отложить «на потом». Поле типа `json.RawMessage` остаётся сырыми байтами:

```go
type Envelope struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

data := []byte(`{"kind":"movie","payload":{"title":"Дюна","year":2021}}`)
var env Envelope
json.Unmarshal(data, &env) // payload пока не разобран

if env.Kind == "movie" {
	var m Movie
	json.Unmarshal(env.Payload, &m)
	fmt.Println(m.Title, m.Year) // Дюна 2021
}
```

Классика для сообщений вида «тип + произвольное тело»: сначала узнаём `kind`, потом разбираем `payload` в подходящую структуру. Без `RawMessage` пришлось бы тащить всё через `map[string]any`.

## Типичные ошибки новичка

| Код | Что случится | Как чинить |
|---|---|---|
| `json.Unmarshal(data, v)` без `&` | ошибка «non-pointer», данные никуда не запишутся | передавать указатель: `&v` |
| поле `score int` с маленькой буквы | молча пропадёт из JSON | экспортировать: `Score int` + тег `` `json:"score"` `` |
| `` `json:"omitempty"` `` без запятой | поле уедет под именем `omitempty` | опция — после запятой: `` `json:"score,omitempty"` `` |
| `omitempty` на вложенной структуре | не сработает: структура не бывает «пустой» | поле-указатель `*Address` — nil исчезнет |
| `v["n"].(int)` для числа из `map[string]any` | паника: JSON-числа — всегда `float64` | `int(v["n"].(float64))` |
| не проверить `err` у `Decode`/`Unmarshal` | кривой JSON тихо оставит нулевые значения | `if err != nil` — каждый раз |
| `io.ReadAll` + `Unmarshal` для файла-гиганта | весь файл в памяти | `json.NewDecoder(f).Decode(&v)` |

## Запомнить

- База: `Marshal` → `[]byte` + ошибка, `Unmarshal(data, &v)` — только указатель; видны лишь экспортированные поля; имена задают теги `` `json:"name"` ``.
- `omitempty` пропускает пустое (`0`, `""`, `false`, `nil`, слайс/карту нулевой длины; структуру не спрячет — нужен указатель); `` `json:"-"` `` — поля не существует; `` `json:"omitempty"` `` без запятой — ловушка-переименование.
- Поле-указатель различает «нет/`null`» (`nil`) и «настоящий ноль» (указатель на 0). Отсутствующие в JSON поля `Unmarshal` не трогает — так делают умолчания.
- Потоки: `json.NewDecoder(r).Decode(&v)` и `json.NewEncoder(w).Encode(v)` — файл, сеть, stdin; `Encode` дописывает `\n`.
- NDJSON читают циклом `for dec.More() { dec.Decode(&e) }` — один декодер, много значений.
- Неизвестная форма — `map[string]any` + type switch; числа там всегда `float64`: `int(x.(float64))`.
- `json.RawMessage` откладывает разбор куска JSON на потом.
