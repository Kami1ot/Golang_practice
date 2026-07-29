package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// fsSnapshot — «транзакция» над файлами контента: запоминает старое состояние
// затрагиваемых файлов и позволяет откатить запись, если валидация провалилась.
// Удаляемые ПАПКИ переезжают в <contentDir>/.trash/ (вне courses/ — иначе
// загрузчик их подберёт) и возвращаются оттуда при откате.
type fsSnapshot struct {
	contentDir string
	files      map[string][]byte // путь -> старые байты; nil = файла не было
	movedDirs  map[string]string // исходный путь -> путь в .trash
}

func newSnapshot(contentDir string) *fsSnapshot {
	return &fsSnapshot{
		contentDir: contentDir,
		files:      map[string][]byte{},
		movedDirs:  map[string]string{},
	}
}

// rememberFile фиксирует текущее содержимое файла (или его отсутствие).
func (s *fsSnapshot) rememberFile(path string) error {
	if _, ok := s.files[path]; ok {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.files[path] = nil
			return nil
		}
		return err
	}
	s.files[path] = data
	return nil
}

// moveDirToTrash убирает папку в .trash (для удалений и «перезаписей»).
func (s *fsSnapshot) moveDirToTrash(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	trashRoot := filepath.Join(s.contentDir, ".trash")
	if err := os.MkdirAll(trashRoot, 0o755); err != nil {
		return err
	}
	buf := make([]byte, 8)
	rand.Read(buf)
	dest := filepath.Join(trashRoot, hex.EncodeToString(buf)+"-"+filepath.Base(dir))
	if err := os.Rename(dir, dest); err != nil {
		return fmt.Errorf("перемещение %s в корзину: %w", dir, err)
	}
	s.movedDirs[dir] = dest
	return nil
}

// restore возвращает всё как было (вызывается при провале валидации).
func (s *fsSnapshot) restore() {
	for path, old := range s.files {
		if old == nil {
			os.Remove(path)
			continue
		}
		if err := os.WriteFile(path, old, 0o644); err != nil {
			log.Printf("откат %s: %v", path, err)
		}
	}
	for orig, trashed := range s.movedDirs {
		os.RemoveAll(orig) // на случай, если успели записать новое
		if err := os.Rename(trashed, orig); err != nil {
			log.Printf("откат папки %s: %v", orig, err)
		}
	}
}

// discard подтверждает изменения и чистит корзину (Windows может держать
// хэндлы — ретраи, как в runner).
func (s *fsSnapshot) discard() {
	for _, trashed := range s.movedDirs {
		go removeAllRetry(trashed)
	}
}

func removeAllRetry(dir string) {
	delays := []time.Duration{0, 200 * time.Millisecond, 500 * time.Millisecond, time.Second}
	var err error
	for _, d := range delays {
		time.Sleep(d)
		if err = os.RemoveAll(dir); err == nil {
			return
		}
	}
	log.Printf("не удалось удалить %s: %v", dir, err)
}

// sweepTrash подметает корзину при старте (брошенные хвосты аварийных откатов).
func SweepTrash(contentDir string) {
	trashRoot := filepath.Join(contentDir, ".trash")
	entries, err := os.ReadDir(trashRoot)
	if err != nil {
		return
	}
	for _, e := range entries {
		os.RemoveAll(filepath.Join(trashRoot, e.Name()))
	}
}
