// Package runner компилирует и запускает код пользователя.
//
// Принципиально: сначала `go build -o prog.exe`, затем запуск бинарника
// напрямую — `go run` на Windows нельзя надёжно убить по таймауту
// (осиротевший дочерний процесс), а отдельная сборка ещё и переиспользуется
// на все тест-кейсы.
package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	buildTimeout = 30 * time.Second
	outputCap    = 256 * 1024 // байт на поток; сверх — процесс убивается
	waitDelay    = 2 * time.Second
)

// Статусы результата (общие для io и unittest).
const (
	StatusPassed       = "passed"
	StatusFailed       = "failed"
	StatusCompileError = "compile_error"
	StatusTimeout      = "timeout"
	StatusRuntimeError = "runtime_error"
)

type Result struct {
	Status        string
	CompileOutput string // при compile_error, с вычищенными путями temp-папки
	Tests         []TestResult
}

type TestResult struct {
	Name       string
	Passed     bool
	Hidden     bool
	Skipped    bool // не запускался (предыдущий тест превысил время)
	Stdin      string
	Expected   string
	Actual     string
	Output     string // stderr / сообщения go test
	DurationMs int64
}

type Runner struct {
	goBin     string
	goModLine string        // строка "go 1.26" для go.mod рабочей области
	sem       chan struct{} // один запуск за раз
}

// New находит go-тулчейн (PATH, затем стандартный путь установки Windows)
// и подметает брошенные рабочие папки прошлых запусков.
func New() (*Runner, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		fallback := `C:\Program Files\Go\bin\go.exe`
		if _, statErr := os.Stat(fallback); statErr == nil {
			goBin = fallback
		} else {
			return nil, fmt.Errorf("go не найден ни в PATH, ни в %s — установите Go", fallback)
		}
	}

	r := &Runner{goBin: goBin, sem: make(chan struct{}, 1)}
	if err := r.detectGoVersion(); err != nil {
		return nil, err
	}
	r.sweepStale()
	return r, nil
}

func (r *Runner) detectGoVersion() error {
	out, err := exec.Command(r.goBin, "env", "GOVERSION").Output()
	if err != nil {
		return fmt.Errorf("go env GOVERSION: %w", err)
	}
	v := strings.TrimSpace(string(out)) // например "go1.26.5"
	v = strings.TrimPrefix(v, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		v = parts[0] + "." + parts[1]
	}
	r.goModLine = "go " + v
	return nil
}

// PreWarm прогревает GOCACHE фоновой сборкой hello-world: первая компиляция
// после установки Go занимает несколько секунд, не нужно платить их
// на первом запуске задачи пользователем.
func (r *Runner) PreWarm() {
	go func() {
		ws, err := r.makeWorkspace(map[string]string{
			"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"warm\") }\n",
		})
		if err != nil {
			log.Printf("pre-warm: %v", err)
			return
		}
		defer removeAllRetry(ws)
		ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
		defer cancel()
		start := time.Now()
		if _, err := r.build(ctx, ws); err != nil {
			log.Printf("pre-warm сборка: %v", err)
			return
		}
		log.Printf("pre-warm: GOCACHE прогрет за %s", time.Since(start).Round(time.Millisecond))
	}()
}

// runEnv — окружение для сборок/запусков кода пользователя.
// Наследуем всё (GOCACHE живёт между запусками), сеть запрещаем.
func runEnv() []string {
	return append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=mod")
}

func (r *Runner) makeWorkspace(files map[string]string) (string, error) {
	ws, err := os.MkdirTemp("", "gotask-")
	if err != nil {
		return "", fmt.Errorf("временная папка: %w", err)
	}
	files["go.mod"] = "module usercode\n\n" + r.goModLine + "\n"
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(ws, name), []byte(content), 0o644); err != nil {
			removeAllRetry(ws)
			return "", fmt.Errorf("запись %s: %w", name, err)
		}
	}
	return ws, nil
}

// build компилирует пакет рабочей области в prog.exe.
// Возвращает вывод компилятора при ошибке сборки (nil error, ненулевой output).
func (r *Runner) build(ctx context.Context, ws string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, buildTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.goBin, "build", "-o", "prog.exe", ".")
	cmd.Dir = ws
	cmd.Env = runEnv()
	cmd.WaitDelay = waitDelay
	out, err := cmd.CombinedOutput()
	if err == nil {
		return "", nil
	}
	if ctx.Err() != nil {
		return "", fmt.Errorf("сборка превысила %s", buildTimeout)
	}
	return cleanPaths(string(out), ws), nil
}

// cleanPaths убирает путь временной папки из сообщений компилятора,
// чтобы пользователь видел main.go:5:2, а не C:\...\gotask-123\main.go:5:2.
func cleanPaths(s, ws string) string {
	s = strings.ReplaceAll(s, ws+string(filepath.Separator), "")
	s = strings.ReplaceAll(s, ws, "")
	s = strings.ReplaceAll(s, "# usercode\n", "")
	return strings.TrimSpace(s)
}

// cappedWriter копит до limit байт, дальше отбрасывает и зовёт onOverflow
// (один раз) — так бесконечный fmt.Println не съедает память: процесс
// убивается отменой контекста.
type cappedWriter struct {
	buf        strings.Builder
	limit      int
	overflowed bool
	onOverflow func()
	mu         sync.Mutex
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	room := w.limit - w.buf.Len()
	if room > 0 {
		if len(p) < room {
			room = len(p)
		}
		w.buf.WriteString(string(p[:room]))
	}
	if w.buf.Len() >= w.limit && !w.overflowed {
		w.overflowed = true
		if w.onOverflow != nil {
			w.onOverflow()
		}
	}
	return len(p), nil // всегда «успех»: усечение не должно ломать процесс раньше kill
}

func (w *cappedWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func (w *cappedWriter) Overflowed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.overflowed
}

// removeAllRetry удаляет папку с повторами: Windows (антивирус, поздний
// teardown процесса) может недолго держать хэндл на prog.exe.
func removeAllRetry(dir string) {
	delays := []time.Duration{0, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}
	var err error
	for _, d := range delays {
		time.Sleep(d)
		if err = os.RemoveAll(dir); err == nil {
			return
		}
	}
	log.Printf("не удалось удалить %s: %v (будет подметено при следующем старте)", dir, err)
}

// sweepStale удаляет рабочие папки старше часа, брошенные из-за
// незавершившихся удалений или аварийных остановок сервера.
func (r *Runner) sweepStale() {
	pattern := filepath.Join(os.TempDir(), "gotask-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Hour)
	for _, dir := range matches {
		info, err := os.Stat(dir)
		if err == nil && info.ModTime().Before(cutoff) {
			os.RemoveAll(dir)
		}
	}
}
