package storage

import (
	"context"
	"errors"
	"jrn/internal/domain"
	"os"
	"path/filepath"
	"time"
)

type FS struct {
	dir string
}

// store, err := storage.New("~/.jrn")
func New(dir string) (*FS, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FS{dir: dir}, nil
}

// err := store.Save(ctx, date, markdownBytes)
func (s *FS) Save(ctx context.Context, date time.Time, data []byte) error {
	filePath := s.datePath(date)
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "jrn_tmp_*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, filePath); err != nil {
		return err
	}

	tmpName = ""
	return nil
}

// data, err := store.Load(ctx, date)
func (s *FS) Load(ctx context.Context, date time.Time) ([]byte, error) {
	data, err := os.ReadFile(s.datePath(date))
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrNotFound
	}
	return data, err
}

func (s *FS) FindPreviousDate(ctx context.Context, before time.Time) (time.Time, bool, error) {
	years, err := os.ReadDir(s.dir)
	if err != nil {
		return time.Time{}, false, err
	}

	for i := len(years) - 1; i >= 0; i-- {
		year := years[i]
		if !year.IsDir() {
			continue
		}

		if year.Name() > before.Format("2006") {
			continue
		}

		months, err := os.ReadDir(filepath.Join(s.dir, year.Name()))
		if err != nil {
			return time.Time{}, false, err
		}

		for j := len(months) - 1; j >= 0; j-- {
			month := months[j]
			if !month.IsDir() {
				continue
			}

			if year.Name() == before.Format("2006") && month.Name() > before.Format("01") {
				continue
			}

			days, err := os.ReadDir(filepath.Join(s.dir, year.Name(), month.Name()))
			if err != nil {
				return time.Time{}, false, err
			}

			for k := len(days) - 1; k >= 0; k-- {
				day := days[k]
				if day.IsDir() {
					continue
				}

				if day.Name() >= before.Format("2006-01-02")+".md" {
					continue
				}
				filename := day.Name()

				if len(filename) < 13 || filename[10:] != ".md" {
					continue
				}

				if !isValidDateStr(filename[:10]) {
					continue
				}
				t, err := time.Parse("2006-01-02", filename[:10])
				if err != nil {
					return time.Time{}, false, err
				}
				return t, true, nil
			}
		}
	}
	return time.Time{}, false, nil
}

func (s *FS) FindNextDate(ctx context.Context, after time.Time) (time.Time, bool, error) {
	years, err := os.ReadDir(s.dir)
	if err != nil {
		return time.Time{}, false, err
	}

	for i := 0; i < len(years); i++ {
		year := years[i]
		if !year.IsDir() {
			continue
		}

		if year.Name() < after.Format("2006") {
			continue
		}

		months, err := os.ReadDir(filepath.Join(s.dir, year.Name()))
		if err != nil {
			return time.Time{}, false, err
		}

		for j := 0; j < len(months); j++ {
			month := months[j]
			if !month.IsDir() {
				continue
			}

			if year.Name() == after.Format("2006") && month.Name() < after.Format("01") {
				continue
			}

			days, err := os.ReadDir(filepath.Join(s.dir, year.Name(), month.Name()))
			if err != nil {
				return time.Time{}, false, err
			}

			for k := 0; k < len(days); k++ {
				day := days[k]
				if day.IsDir() {
					continue
				}

				if day.Name() <= after.Format("2006-01-02")+".md" {
					continue
				}
				filename := day.Name()

				if len(filename) < 13 || filename[10:] != ".md" {
					continue
				}

				if !isValidDateStr(filename[:10]) {
					continue
				}
				t, err := time.Parse("2006-01-02", filename[:10])
				if err != nil {
					return time.Time{}, false, err
				}
				return t, true, nil
			}
		}
	}
	return time.Time{}, false, nil
}

func (s *FS) ListDates(ctx context.Context, date time.Time) ([]time.Time, error) {
	var dates []time.Time
	years, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	for _, year := range years {
		if !year.IsDir() {
			continue
		}

		months, err := os.ReadDir(filepath.Join(s.dir, year.Name()))
		if err != nil {
			return nil, err
		}

		for _, month := range months {
			if !month.IsDir() {
				continue
			}

			days, err := os.ReadDir(filepath.Join(s.dir, year.Name(), month.Name()))
			if err != nil {
				return nil, err
			}

			for _, day := range days {
				if day.IsDir() {
					continue
				}
				filename := day.Name()

				if len(filename) < 13 || filename[10:] != ".md" {
					continue
				}

				if !isValidDateStr(filename[:10]) {
					continue
				}

				if day.Name() >= date.Format("2006-01-02")+".md" {
					continue
				}

				t, err := time.Parse("2006-01-02", filename[:10])
				if err != nil {
					continue
				}
				dates = append(dates, t)
			}
		}
	}

	return dates, nil
}

func (s *FS) Exists(ctx context.Context, date time.Time) (bool, error) {
	_, err := os.Stat(s.datePath(date))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func isValidDateStr(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	for i, ch := range []byte(s) {
		if i == 4 || i == 7 {
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func (s *FS) datePath(date time.Time) string {
	year := date.Format("2006")
	month := date.Format("01")
	filename := date.Format("2006-01-02") + ".md"
	return filepath.Join(s.dir, year, month, filename)
}
