package ocr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrLibraryNotFound is returned by FindLibrary when no OCR shared library
// can be located in the search path.
var ErrLibraryNotFound = errors.New("ocr: native library not found")

// librarySearchDirs returns the directories probed for the native library:
// OCR_LIBRARY_PATH first (if set), then the executable directory, the current
// working directory, and the OS-specific default locations.
func librarySearchDirs() []string {
	var dirs []string

	if p := os.Getenv("OCR_LIBRARY_PATH"); p != "" {
		for _, d := range filepath.SplitList(p) {
			if d != "" {
				dirs = append(dirs, d)
			}
		}
	}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}
	dirs = append(dirs, systemLibraryDirs()...)

	// De-duplicate while preserving order.
	seen := make(map[string]bool, len(dirs))
	out := dirs[:0]
	for _, d := range dirs {
		d = filepath.Clean(d)
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// FindLibrary looks for the native OCR shared library and returns the path
// of the first existing candidate. The search honors the OCR_LIBRARY_PATH
// environment variable, then checks the executable directory, current
// working directory and standard OS locations, trying every naming variant
// for the current operating system (libocr.so / libocr.dylib / ocr.dll and
// their versioned names).
func FindLibrary() (string, error) {
	dirs := librarySearchDirs()
	names := defaultLibraryNames()

	var tried []string
	for _, d := range dirs {
		for _, name := range names {
			candidate := filepath.Join(d, name)
			tried = append(tried, candidate)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("%w for %s/%s; searched:\n  %s\n"+
		"install the library, set OCR_LIBRARY_PATH, or put it next to the executable",
		ErrLibraryNotFound, runtime.GOOS, runtime.GOARCH, strings.Join(tried, "\n  "))
}
