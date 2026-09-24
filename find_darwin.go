//go:build darwin

package ocr

func systemLibraryDirs() []string {
	return []string{
		"/usr/local/lib",
		"/opt/homebrew/lib",
		"/opt/local/lib",
		"/usr/lib",
	}
}
