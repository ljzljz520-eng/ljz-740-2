//go:build windows

package ocr

// On Windows the runtime artifact is a DLL; no "lib" prefix and a .dll suffix.
func defaultLibraryNamesOS() []string {
	return []string{"ocr.dll", "libocr.dll"}
}
