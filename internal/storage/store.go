package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func DataDirectory(appID string) (string, error) {
	var dir string
	var err error
	switch runtime.GOOS {
	case "linux":
		dir = os.Getenv("XDG_DATA_HOME")
		if !filepath.IsAbs(dir) {
			dir, err = os.UserHomeDir()
			dir = filepath.Join(dir, ".local", "share")
		}
	default:
		dir, err = os.UserConfigDir()
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appID), nil
}

type Store struct {
	dir string
	mu  sync.Mutex
}

// New returns a store rooted in the application's private data directory.
func New(dir string) *Store { return &Store{dir: dir} }

// Dir returns the root directory for native file operations.
func (s *Store) Dir() string { return s.dir }

func Load[T any](s *Store, name string, defaults func() T) (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults(), nil
	}
	if err != nil {
		return defaults(), err
	}
	value := defaults()
	if err := json.Unmarshal(data, &value); err != nil {
		backup := fmt.Sprintf("%s.corrupt-%d", path, time.Now().UnixNano())
		if renameErr := os.Rename(path, backup); renameErr != nil {
			return defaults(), fmt.Errorf("backup %s: %w", name, renameErr)
		}
		log.Printf("%s: %v (backed up to %s)", name, err, backup)
		return defaults(), nil
	}
	return value, nil
}

func Save(s *Store, name string, value any, maxBytes int) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if maxBytes > 0 && len(data) > maxBytes {
		return errors.New("JSON exceeds the file size limit")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, writeErr := tmp.Write(append(data, '\n'))
	syncErr := tmp.Sync()
	if err := errors.Join(writeErr, syncErr, tmp.Close()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func ValidateFilename(name string) error {
	if name == "" || len(name) > 200 || strings.HasPrefix(name, ".") ||
		strings.Contains(name, "..") || strings.ContainsAny(name, "/\\:\x00") {
		return errors.New("invalid filename")
	}
	return nil
}

func ReadLimited(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("expected a regular file within the size limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("file exceeds the size limit")
	}
	return data, nil
}

func CleanupReports(dir, extension string, olderThan time.Duration) (int, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var removeErr error
	removed := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != extension || !entry.Type().IsRegular() {
			continue
		}
		if olderThan > 0 {
			info, err := entry.Info()
			if err != nil || time.Since(info.ModTime()) <= olderThan {
				continue
			}
		}
		err := os.Remove(filepath.Join(dir, entry.Name()))
		if err == nil {
			removed++
		} else if !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return removed, removeErr
}
