package main

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessBasic(t *testing.T) {
	got := Process([]int{1, 2, 3, 4, 5}, 2, func(n int) int { return n * 2 })
	want := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Process = %v, ожидалось %v — результаты в исходном порядке", got, want)
	}
}

func TestProcessMoreWorkersThanJobs(t *testing.T) {
	done := make(chan []int, 1)
	go func() {
		done <- Process([]int{7}, 10, func(n int) int { return n * 2 })
	}()
	select {
	case got := <-done:
		if !reflect.DeepEqual(got, []int{14}) {
			t.Errorf("Process([7], 10 воркеров) = %v, ожидалось [14]", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Process завис при workers > числа задач — лишние воркеры должны просто завершиться (close канала задач!)")
	}
}

func TestProcessEmpty(t *testing.T) {
	got := Process(nil, 3, func(n int) int { return n })
	if len(got) != 0 {
		t.Errorf("Process(nil) = %v, ожидался пустой результат", got)
	}
}

func TestProcessLimitsConcurrency(t *testing.T) {
	const workers = 3
	var current, peak int64
	var mu sync.Mutex
	f := func(n int) int {
		cur := atomic.AddInt64(&current, 1)
		mu.Lock()
		if cur > peak {
			peak = cur
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond) // подержим задачу, чтобы поймать параллельность
		atomic.AddInt64(&current, -1)
		return n
	}
	nums := make([]int, 30)
	for i := range nums {
		nums[i] = i
	}
	Process(nums, workers, f)
	mu.Lock()
	defer mu.Unlock()
	if peak > workers {
		t.Errorf("одновременно работало %d задач при лимите %d — пул должен ограничивать параллельность, а не запускать горутину на элемент", peak, workers)
	}
	if peak == 0 {
		t.Errorf("не зафиксировано ни одной работающей задачи — f вообще вызывалась?")
	}
}

func TestProcessActuallyConcurrent(t *testing.T) {
	// Барьер: первые две задачи должны работать ОДНОВРЕМЕННО.
	// Последовательное выполнение застрянет на первой — это ловит сторожевой таймер.
	ready := make(chan struct{}, 8)
	release := make(chan struct{})
	var once sync.Once
	f := func(n int) int {
		ready <- struct{}{}
		<-release
		return n
	}
	go func() {
		<-ready
		<-ready // дождались двух одновременных задач
		once.Do(func() { close(release) })
	}()

	done := make(chan struct{})
	go func() {
		Process([]int{1, 2, 3, 4, 5, 6}, 3, f)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("Process выполняет задачи по одной — нужен настоящий пул из workers горутин, а не последовательный цикл")
	}
}

func TestProcessOrderWithUnevenWork(t *testing.T) {
	// разное «время выполнения» — порядок всё равно исходный, слоты решают
	f := func(n int) int {
		if n%2 == 0 {
			time.Sleep(3 * time.Millisecond)
		}
		return n * n
	}
	nums := []int{5, 4, 3, 2, 1}
	got := Process(nums, 2, f)
	want := []int{25, 16, 9, 4, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Process с неравной работой = %v, ожидалось %v", got, want)
	}
}
