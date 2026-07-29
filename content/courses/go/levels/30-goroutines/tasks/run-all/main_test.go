package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAllCompletes(t *testing.T) {
	var counter int64
	funcs := make([]func(), 50)
	for i := range funcs {
		funcs[i] = func() { atomic.AddInt64(&counter, 1) }
	}
	RunAll(funcs)
	if got := atomic.LoadInt64(&counter); got != 50 {
		t.Errorf("к моменту возврата RunAll выполнилось %d функций из 50 — Wait должен дождаться всех", got)
	}
}

func TestRunAllConcurrent(t *testing.T) {
	// Барьер: каждая функция ждёт, пока стартуют ВСЕ.
	// В горутинах барьер проходится мгновенно; последовательный цикл
	// застрянет на первой же функции — это и ловит сторожевой таймер.
	const n = 8
	ready := make(chan struct{}, n)
	release := make(chan struct{})
	funcs := make([]func(), n)
	for i := range funcs {
		funcs[i] = func() {
			ready <- struct{}{}
			<-release
		}
	}
	go func() {
		for i := 0; i < n; i++ {
			<-ready
		}
		close(release)
	}()

	done := make(chan struct{})
	go func() {
		RunAll(funcs)
		close(done)
	}()
	select {
	case <-done:
		// успех: все 8 функций работали одновременно
	case <-time.After(3 * time.Second):
		t.Fatalf("RunAll застряла: функции выполняются НЕ конкурентно — каждая должна работать в своей горутине")
	}
}

func TestRunAllEmpty(t *testing.T) {
	done := make(chan struct{})
	go func() {
		RunAll(nil)
		RunAll([]func(){})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("RunAll пустого слайса зависла — Wait ждёт дел, которых нет")
	}
}
