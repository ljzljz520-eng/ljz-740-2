//go:build !linux && !darwin && !windows

package ocr

// Best-effort POSIX-style fallback for other Unix systems (BSD, Solaris...).
func defaultLibraryNamesOS() []string {
	return []string{"libocr.so"}
}
