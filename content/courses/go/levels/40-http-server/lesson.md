# net/http: сервер — вы по ту сторону запроса

На прошлом уровне вы писали клиента: собирали запрос, отправляли, разбирали статус и тело ответа. Сегодня пересаживаемся на другой конец провода — будем эти ответы *формировать*. Хорошая новость: сервер в Go — не фреймворк на тысячу зависимостей, а пара типов из стандартного пакета `net/http`. Платформа GoPractice, в которой вы читаете этот урок, — ровно такой сервер: один Go-бинарник с хендлерами.

## Хендлер — функция, которой отдают запрос

Сервер собирается из обработчиков (*handler*) — обычных функций с фиксированной сигнатурой:

```go
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Привет из хендлера!")
}
```

Два параметра — и весь HTTP у вас в руках:

- `r *http.Request` — всё о запросе: метод, путь, параметры, заголовки, тело. Тот самый тип, который вы собирали на уровне про клиента, — теперь вы его *получаете*.
- `w http.ResponseWriter` — «куда писать ответ». Присмотритесь: у него есть метод `Write([]byte) (int, error)` — это же `io.Writer` с уровня про потоки! Поэтому работает вся знакомая артиллерия: `fmt.Fprintf(w, ...)`, `fmt.Fprintln(w, ...)`, `json.NewEncoder(w)`.

Написали в `w` — байты уехали клиенту. Никакой магии.

## ServeMux: кто какой путь обслуживает

Хендлеров в сервере много, и кто-то должен раздавать им запросы. Это маршрутизатор (*router*), в стандартной библиотеке — `http.ServeMux`:

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/bye", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Пока!")
	})

	err := http.ListenAndServe("127.0.0.1:8080", mux)
	fmt.Println("сервер упал:", err)
}
```

`ListenAndServe` занимает порт и крутится вечно: принимает соединения и передаёт каждый запрос хендлеру с подходящим шаблоном; путь без хендлера получает 404. Проверить можно знакомым curl:

```
$ curl http://127.0.0.1:8080/hello
Привет из хендлера!
```

> [!NOTE]
> `ListenAndServe` возвращает управление только при ошибке — программа «висит» в нём намеренно, это и есть работа сервера. В задачах этого уровня запускать сервер НЕ нужно: хендлеры проверяются напрямую, без порта, — как именно, увидите в конце урока.

## Маршруты Go 1.22+: метод и {id} прямо в шаблоне

Современный `ServeMux` понимает в шаблоне HTTP-метод и подстановки (*wildcards*):

```go
mux.HandleFunc("GET /ping", ping)
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("POST /users", createUser)
```

Два бонуса сразу:

- **Метод проверяет сам mux.** На `POST /ping` клиент получит `405 Method Not Allowed` — вы не пишете ни строчки.
- **Кусок пути — в переменную.** `{id}` совпадает с одним сегментом, а в хендлере его отдаёт `r.PathValue("id")` — строкой:

```go
func getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // "42" для запроса /users/42
	fmt.Fprintf(w, "профиль пользователя %s", id)
}
```

В коде, написанном до Go 1.22, вы встретите ручную проверку — она делает то же самое, только руками:

```go
mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // в старом коде — сплошь и рядом
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprint(w, "pong")
})
```

Узнавать этот паттерн полезно, писать так самому — уже не нужно.

## Что пришло в запросе

Всё интересное лежит в `r`:

```go
r.Method                  // "GET", "POST", …
r.URL.Path                // "/users/42"
r.URL.Query().Get("q")    // query-параметр ?q=…; нет параметра — ""
r.PathValue("id")         // подстановка {id} из шаблона маршрута
r.Header.Get("Accept")    // заголовок запроса
r.Body                    // тело: io.Reader — да, тот самый
```

Query-параметры — необязательные уточнения после `?` в URL:

```go
func greet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "гость"
	}
	fmt.Fprintf(w, "Привет, %s!", name)
}
```

```
$ curl "http://127.0.0.1:8080/greet?name=Ада"
Привет, Ада!
$ curl "http://127.0.0.1:8080/greet"
Привет, гость!
```

Тело запроса — `io.Reader`, поэтому JSON из него декодируется ровно как на уровне про JSON — потоком, без промежуточной строки:

```go
func createUser(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "кривой JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "создаём пользователя %s", in.Name)
}
```

> [!NOTE]
> Закрывать `r.Body` в хендлере не нужно — сервер сделает это сам после ответа. Это клиент был обязан писать `defer resp.Body.Close()`; на серверной стороне за вами убирают.

## Ответ: статус, заголовки, тело

Если хендлер просто пишет в `w`, клиент получает `200 OK` — статус по умолчанию. Другой статус нужно объявить **до** первой записи тела:

```go
w.WriteHeader(http.StatusNotFound) // 404 — и только потом тело
fmt.Fprintln(w, "нет такой страницы")
```

Для пары «статус + текст ошибки» есть готовый помощник — им пользуются чаще, чем голым `WriteHeader`:

```go
http.Error(w, "нет такого пользователя", http.StatusNotFound)
return // не забудьте выйти: http.Error за вас этого не сделает!
```

Заголовки ответа складывают в `w.Header()` — тоже строго до `WriteHeader` и записи тела. Полный JSON-ответ выглядит так:

```go
func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id должен быть числом", http.StatusBadRequest)
		return
	}
	u, ok := users[id]
	if !ok {
		http.Error(w, "нет такого пользователя", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
```

```
$ curl -i http://127.0.0.1:8080/users/1
HTTP/1.1 200 OK
Content-Type: application/json

{"id":1,"name":"Ада"}
```

Запомните порядок — он железный: **заголовки → статус → тело**. HTTP-ответ уезжает клиенту в этом порядке физически, поэтому «передумать» после начала записи нельзя.

> [!WARNING]
> Самая частая пара ошибок новичка: `WriteHeader` после записи тела — в лог падает `superfluous response.WriteHeader`, а клиент уже получил 200; и забытый `return` после `http.Error` — хендлер продолжает выполняться и дописывает в тело ответа то, чего там быть не должно.

## Каждый запрос — своя горутина

Сервер обрабатывает запросы конкурентно: на каждый — своя горутина. Для вас это бесплатная производительность и знакомая по секции о горутинах ответственность: два запроса могут **одновременно** трогать общие данные. Общее состояние — под мьютекс:

```go
var (
	mu    sync.Mutex
	hits  int
)

func stats(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	hits++
	n := hits
	mu.Unlock()
	fmt.Fprintf(w, "вы посетитель №%d", n)
}
```

> [!NOTE]
> Паника в хендлере НЕ роняет сервер: сгорает только горутина этого запроса — клиент получает оборванный ответ, в лог падает стек, остальные запросы живут. Но полагаться на это не стоит: паника — всё ещё авария, а не обработка ошибок.

## Тестируем хендлер без сервера

Как проверить хендлер, не занимая порт? Пакет `net/http/httptest` умеет подделывать обе стороны:

```go
req := httptest.NewRequest("GET", "/greet?name=Ада", nil) // готовый *http.Request
rec := httptest.NewRecorder() // ResponseWriter-магнитофон: запишет всё, что выведет хендлер

greet(rec, req) // хендлер — обычная функция: просто вызываем

fmt.Println(rec.Code)          // 200
fmt.Println(rec.Body.String()) // Привет, Ада!
```

`rec.Code` — статус, `rec.Body` — тело, `rec.Header()` — заголовки ответа. Целый маршрутизатор проверяется так же — у него есть метод `ServeHTTP`:

```go
rec := httptest.NewRecorder()
mux.ServeHTTP(rec, httptest.NewRequest("POST", "/ping", nil))
fmt.Println(rec.Code) // 405 — метод не совпал с шаблоном "GET /ping"
```

Именно так проверяются задачи этого уровня — ни одного занятого порта. А `httptest.NewServer` с прошлого уровня — второй этаж той же идеи: настоящий сервер на случайном порту, когда нужно прогнать запрос через реальную сеть.

## Типичные ошибки новичка

| Код | Что случится | Как чинить |
|---|---|---|
| `fmt.Fprint(w, ...)`, потом `w.WriteHeader(404)` | в логе `superfluous response.WriteHeader`, клиент получил 200 | сначала статус, потом тело |
| `w.Header().Set(...)` после `WriteHeader` | заголовок молча теряется | заголовки — до статуса и тела |
| `http.Error(w, ...)` без `return` | хендлер выполняется дальше и дописывает тело | `return` сразу после |
| общая мапа/счётчик из хендлеров без мьютекса | гонка данных: запросы конкурентны | `sync.Mutex`, как в секции про горутины |
| шаблон `/ping` без метода | POST и DELETE тоже проходят в хендлер | метод в шаблоне: `"GET /ping"` |
| проверка хендлеров запуском `ListenAndServe` | тест занимает порт и висит вечно | `httptest.NewRequest` + `NewRecorder` |

## Запомнить

- Хендлер — функция `func(w http.ResponseWriter, r *http.Request)`; `w` — это `io.Writer`, поэтому `fmt.Fprintf` и `json.NewEncoder(w).Encode` работают из коробки.
- `mux := http.NewServeMux()`, `mux.HandleFunc(шаблон, хендлер)`, `http.ListenAndServe(адрес, mux)` — весь сервер.
- Шаблоны Go 1.22+: `"GET /users/{id}"` — метод проверяет mux (иначе 405), сегмент отдаёт `r.PathValue("id")`.
- Данные запроса: `r.URL.Query().Get("name")` — query-параметры, `r.Body` + `json.NewDecoder` — JSON-тело.
- Порядок ответа железный: `w.Header().Set(...)` → `w.WriteHeader(код)` → тело; по умолчанию статус 200; `http.Error(w, текст, код)` + `return` — стандартная пара для ошибок.
- Каждый запрос — своя горутина: общее состояние защищайте мьютексом.
- Тесты без порта: `httptest.NewRequest` + `httptest.NewRecorder` → вызвать хендлер (или `mux.ServeHTTP`) → проверить `rec.Code`, `rec.Body`.
