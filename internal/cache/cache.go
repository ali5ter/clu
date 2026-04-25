// Package cache persists API responses to ~/.cache/clu/ for offline use.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func dir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(base, "clu")
	return d, os.MkdirAll(d, 0o755)
}

// Write serialises v to <cacheDir>/<name>.json.
func Write(name string, v any) error {
	d, err := dir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, name+".json"), data, 0o644)
}

// Read deserialises <cacheDir>/<name>.json into dst.
func Read(name string, dst any) error {
	d, err := dir()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(d, name+".json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
