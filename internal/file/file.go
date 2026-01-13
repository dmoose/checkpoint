package file

import (
	"fmt"
	"os"
)

func ReadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(b), nil
}

func WriteFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FindWithFallback checks for a file at the primary path, then falls back to legacy path.
// Returns the path that exists, or the primary path if neither exists.
func FindWithFallback(primary, legacy string) string {
	if Exists(primary) {
		return primary
	}
	if Exists(legacy) {
		return legacy
	}
	return primary // Return primary for creation
}
