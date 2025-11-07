package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteExists(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.txt")
	if Exists(p) {
		t.Fatalf("file should not exist yet")
	}
	if err := WriteFile(p, "hello"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !Exists(p) {
		t.Fatalf("file should exist after write")
	}
	got, err := ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
	// ensure file perms are reasonable (rw for owner)
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Mode()&0600 == 0 {
		t.Fatalf("unexpected file mode: %v", fi.Mode())
	}
}
