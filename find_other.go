//go:build !linux && !darwin && !windows

package ocr

func systemLibraryDirs() []string {
	return []string{"/usr/local/lib", "/usr/lib"}
}
