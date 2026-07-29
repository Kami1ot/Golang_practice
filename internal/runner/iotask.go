package runner

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopractice/internal/content"
)

// RunIO компилирует программу пользователя и прогоняет её по тест-кейсам,
// сравнивая нормализованный stdout с ожидаемым.
func (r *Runner) RunIO(ctx context.Context, task *content.Task, code string) (*Result, error) {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()

	ws, err := r.makeWorkspace(map[string]string{"main.go": code})
	if err != nil {
		return nil, err
	}
	defer removeAllRetry(ws)

	compileOut, err := r.build(ctx, ws)
	if err != nil {
		return nil, err
	}
	if compileOut != "" {
		return &Result{Status: StatusCompileError, CompileOutput: compileOut}, nil
	}

	result := &Result{Status: StatusPassed}
	timedOut := false
	for _, tc := range task.Tests {
		tr := TestResult{Name: tc.Name, Hidden: tc.Hidden}
		if timedOut {
			tr.Skipped = true
			tr.Output = "не запускался: предыдущий тест превысил лимит времени"
			result.Tests = append(result.Tests, tr)
			continue
		}

		outcome := r.runOnce(ctx, ws, tc.Stdin, time.Duration(task.Timeout())*time.Second)
		tr.DurationMs = outcome.durationMs
		tr.Stdin = tc.Stdin
		tr.Expected = Normalize(tc.Stdout)
		tr.Actual = Normalize(outcome.stdout)
		switch {
		case outcome.timedOut:
			tr.Output = "превышен лимит времени (" + time.Duration(task.Timeout()*int(time.Second)).String() + ")"
			if outcome.overflowed {
				tr.Output = "слишком большой вывод (более 256 КБ) — программа остановлена"
			}
			result.Status = StatusTimeout
			timedOut = true
		case outcome.runErr != "":
			tr.Output = outcome.runErr
			if result.Status == StatusPassed {
				result.Status = StatusRuntimeError
			}
		case tr.Actual == tr.Expected:
			tr.Passed = true
		default:
			if result.Status == StatusPassed {
				result.Status = StatusFailed
			}
		}
		result.Tests = append(result.Tests, tr)
	}
	return result, nil
}

type runOutcome struct {
	stdout     string
	runErr     string // stderr + описание ошибки при ненулевом коде выхода
	timedOut   bool
	overflowed bool
	durationMs int64
}

// runOnce запускает собранный prog.exe на одном тест-кейсе.
func (r *Runner) runOnce(parent context.Context, ws, stdin string, timeout time.Duration) runOutcome {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	stdout := &cappedWriter{limit: outputCap, onOverflow: cancel}
	stderr := &cappedWriter{limit: outputCap, onOverflow: cancel}

	cmd := exec.CommandContext(ctx, filepath.Join(ws, "prog.exe"))
	cmd.Dir = ws
	cmd.Env = runEnv()
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = waitDelay

	start := time.Now()
	err := cmd.Run()
	outcome := runOutcome{
		stdout:     stdout.String(),
		durationMs: time.Since(start).Milliseconds(),
		overflowed: stdout.Overflowed() || stderr.Overflowed(),
	}
	if ctx.Err() != nil {
		outcome.timedOut = true
		return outcome
	}
	if err != nil {
		var exitErr *exec.ExitError
		msg := strings.TrimSpace(stderr.String())
		if errors.As(err, &exitErr) {
			if msg == "" {
				msg = err.Error()
			}
		} else {
			msg = err.Error()
		}
		outcome.runErr = cleanPaths(msg, ws)
	}
	return outcome
}
