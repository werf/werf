package timestamps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ReadTimestampFile(path string) (time.Time, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, fmt.Errorf("error accessing %q: %w", path, err)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("error reading %q: %w", path, err)
	}

	i, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		os.RemoveAll(path)
		return time.Time{}, nil
	}

	return time.Unix(i, 0), nil
}

func WriteTimestampFile(path string, t time.Time) error {
	timeStr := fmt.Sprintf("%d\n", t.Unix())

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %w", dir, err)
	}

	if err := ioutil.WriteFile(path, []byte(timeStr), 0o644); err != nil {
		return fmt.Errorf("error writing %q: %w", path, err)
	}

	return nil
}

func CheckTimestampFileExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error accessing %q: %w", path, err)
	}
	return true, nil
}
