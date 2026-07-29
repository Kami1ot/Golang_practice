# Идеи: CI/CD через GitHub и мультиязычные курсы

Статус: **идея, не реализовано**. Записано 2026-07-17 по итогам обсуждения.
Факты о коде проверены по текущему состоянию репозитория.

---

## Часть 1. CI/CD через GitHub Actions

### Предпосылки

- Папка ещё не git-репозиторий. Шаг ноль: `git init` + репозиторий на GitHub
  (делает пользователь сам). Перед первым push проверить `.gitignore`
  (временные `cookies.txt`, локальные json-файлы для curl и т.п.).

### Выбор сервиса: GitHub Actions

Раз код живёт на GitHub — сторонние CI (CircleCI, GitLab CI, Jenkins) не дают
ничего сверху, только внешний сервис и лишнюю настройку. Actions даёт:

- запуск на каждый push/PR, статусы прямо в коммитах и PR;
- бесплатно для публичных репозиториев, 2000 мин/мес для приватных — с запасом;
- сервис-контейнеры (Postgres) из коробки.

### CI-пайплайн (на push/PR)

**Job 1 — основной, `ubuntu-latest`** (быстрый и дешёвый):

1. `actions/checkout` + `actions/setup-go` (версия из `go.mod`, кэш включён).
2. `go vet ./...` и `go build ./...` — вложенный `go.mod` в `content/`
   (модуль gopractice-content) уже защищает от попытки собрать учебные
   `starter.go`, в CI это работает так же.
3. `go test ./...` — store-тесты требуют Postgres:

   ```yaml
   services:
     postgres:
       image: postgres:17
       env:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: root
         POSTGRES_DB: gopractice
       ports: ['5432:5432']
       options: >-
         --health-cmd pg_isready --health-interval 5s
         --health-timeout 5s --health-retries 10
   ```

   и `GOPRACTICE_DSN` в env шага с тестами.
4. **Валидация контента** — отдельный шаг, прогоняющий `content.Load`
   по `content/courses`. Это самая ценная проверка для проекта, где курсы —
   файлы данных: ловит битые `course.json`/`level.json`/`task.json` прямо
   в коммите. Реализация: маленький тест в `internal/content`, который грузит
   реальную папку `content/courses` (или мини-команда `cmd/contentcheck`).

**Job 2 — `windows-latest`, только тесты раннера** (`go test ./internal/runner/...`).
Раннер полон Windows-специфики — `prog.exe`, ретраи удаления файлов из-за
антивируса, kill-семантика таймаутов, — и гонять его только на Linux означало бы
не тестировать реальное окружение. Windows-раннеры медленнее, поэтому им — только
пакет runner, без Postgres.

### Релизы (CD для локального приложения)

Деплоить некуда — приложение локальное, поэтому «CD» = автоматические релизы:

- по пушу тега `v*` — job, который собирает `gopractice.exe`
  (`go build -o gopractice.exe ./cmd/server`, windows/amd64),
  упаковывает вместе с `web/` и `content/` в zip и вешает на GitHub Release
  (`softprops/action-gh-release` или `gh release create`);
- позже, если релизов станет много, можно перейти на goreleaser.

Автодеплой на сервер (VPS) — осознанно за скобками: понадобится только для
публичного хостинга, а это отдельная история с изоляцией раннера (см. часть 2).

### Секреты

`ANTHROPIC_API_KEY` в CI **не нужен и не заводится** — тесты от него не зависят,
чат-наставник в CI не участвует.

---

## Часть 2. Курсы на других языках (приоритет: Python, затем JavaScript/Node)

### Что уже языконезависимо (переиспользуется как есть)

- **io-задачи**: stdin → stdout, сравнение через `Normalize`
  (`internal/runner/normalize.go`) — вообще не знают про язык;
- черновики, прогресс, ачивки, DTO, квизы — тоже.

### Где «Go» зашит жёстко (карта мест)

- `internal/runner/runner.go` — `go build -o prog.exe`, workspace c `go.mod`,
  fallback-путь `C:\Program Files\Go\bin\go.exe`, env `GOPROXY=off`/`GOFLAGS`;
- `internal/runner/iotask.go` — имя файла `main.go`;
- `internal/runner/unittest.go` — `solution.go` + `main_test.go`,
  `go test -json`, парсер test2json;
- `internal/content/loader.go` — обязательные файлы `starter.go`/`main_test.go`;
  `DisallowUnknownFields` в декодере (новое поле JSON требует правки модели);
- `internal/httpapi/admin.go` — имена файлов в веб-редакторе;
- `internal/chat` — `mentorSystemPrompt` («помогаешь новичку изучать язык Go»);
- `web/js/views/tasks.js` — CodeMirror mode `text/x-go`, vendor-файл go-mode.

### Дизайн

1. **Поле `language` в `course.json`** (default `"go"`). Язык — свойство курса:
   это соответствует уже принятому принципу «другой язык = отдельный курс».
   Поле добавляется в модель (`internal/content/model.go`) — из-за
   `DisallowUnknownFields` просто дописать его в JSON нельзя.

2. **Рефакторинг раннера в интерфейс Toolchain + реестр по языку.** Эскиз:

   ```go
   type Toolchain interface {
       Lang() string                 // "go", "python", "js"
       Detect() error                // найти тулчейн (LookPath + фоллбеки)
       StarterFile() string          // starter.go / starter.py / starter.js
       Prepare(ws string) error      // go.mod для Go; для остальных — ничего
       BuildCmd(ws string) *exec.Cmd // nil для интерпретируемых
       RunCmd(ws string) *exec.Cmd   // prog.exe | python main.py | node main.js
   }
   ```

   Go — первая реализация, поведение 1-в-1 с текущим (это чистый рефакторинг,
   существующие тесты раннера должны пройти без изменений). Таймауты, кап вывода
   256 КБ, `cappedWriter`, `waitDelay`, семафор, cleanup с ретраями — общий код,
   не дублируется по языкам.

3. **Новый язык стартует только с io-задачами** — они бесплатны при такой схеме.
   Режим `unittest` — отдельная реализация на язык (pytest для Python,
   `node:test` для JS), добавляется позже и только если реально понадобится.

4. **Интерпретируемые языки проще Go**: нет фазы сборки — сразу
   `python main.py` / `node main.js` с тем же таймаутом и капом.
   Node.js в системе сейчас нет — для JS-курса потребуется установка.

5. **Детект тулчейна при старте сервера**: если `python`/`node` не найден —
   сервер НЕ падает, курс остаётся в витрине с плашкой
   «требуется установить Python», задачи заблокированы. Флаг доступности
   едет в DTO курса.

6. **Изоляция — честная фиксация ограничений.** Сейчас проверка решения — это
   локальный процесс с таймаутом и капом вывода; сеть у программы ученика
   не отрезана (у Go — так же: `GOPROXY=off` режет только скачивание модулей
   при сборке). Для локального личного использования — приемлемо.
   Для публичного хостинга обязателен Docker/песочница — записано как
   осознанное ограничение, а не дыра.

7. **Фронт и контент**: mode CodeMirror по языку курса (добавить vendor-файлы
   python/javascript), расширение starter-файла в загрузчике и админ-редакторе,
   язык курса — в `mentorSystemPrompt` наставника.

### Порядок внедрения (когда дойдёт до дела)

1. Рефакторинг Toolchain на одном Go — без новых языков, тесты зелёные.
2. Поле `language` + детект тулчейнов + плашка в витрине.
3. Python: io-задачи, CodeMirror mode, админка, промпт наставника.
4. JavaScript/Node — по той же схеме.
5. unittest-режим для Python (pytest) — по потребности контента.
