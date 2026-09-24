package ocr

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultLibraryNamesForOS(t *testing.T) {
	names := defaultLibraryNames()
	if len(names) == 0 {
		t.Fatal("expected at least one library name candidate")
	}
	var wantPrefix, wantContains string
	switch runtime.GOOS {
	case "linux":
		wantPrefix, wantContains = "libocr", ".so"
	case "darwin":
		wantPrefix, wantContains = "libocr", ".dylib"
	case "windows":
		wantPrefix, wantContains = "ocr", ".dll"
	default:
		wantPrefix, wantContains = "libocr", ".so"
	}
	for i, n := range names {
		if i == 0 {
			if len(n) < len(wantPrefix) || n[:len(wantPrefix)] != wantPrefix {
				t.Errorf("name %q does not start with %q", n, wantPrefix)
			}
			if !strings.Contains(n, wantContains) {
				t.Errorf("name %q does not contain %q", n, wantContains)
			}
		}
	}
}

func TestFindLibraryViaEnvPath(t *testing.T) {
	dir := t.TempDir()
	names := defaultLibraryNames()
	fake := filepath.Join(dir, names[0])
	if err := os.WriteFile(fake, []byte("not-a-real-lib"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OCR_LIBRARY_PATH", dir)
	got, err := FindLibrary()
	if err != nil {
		t.Fatalf("FindLibrary: %v", err)
	}
	if got != fake {
		t.Fatalf("FindLibrary = %q, want %q", got, fake)
	}
}

func TestFindLibraryNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OCR_LIBRARY_PATH", dir)
	// The executable dir, cwd and system dirs may still contain a real
	// lib; if one is found there, skip instead of asserting.
	_, err := FindLibrary()
	if err == nil {
		t.Skip("an OCR library exists outside OCR_LIBRARY_PATH; skipping not-found case")
	}
	if !errors.Is(err, ErrLibraryNotFound) {
		t.Fatalf("error = %v, want ErrLibraryNotFound", err)
	}
}

func TestLibrarySearchDirsDedup(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OCR_LIBRARY_PATH", dir+string(os.PathListSeparator)+dir)
	dirs := librarySearchDirs()
	seen := map[string]int{}
	for _, d := range dirs {
		seen[d]++
	}
	if seen[filepath.Clean(dir)] != 1 {
		t.Fatalf("env dir counted %d times, want 1 (dirs=%v)", seen[filepath.Clean(dir)], dirs)
	}
}
