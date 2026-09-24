//go:build darwin

package ocr

// On macOS a Mach-O dylib uses the .dylib suffix (plus optional versioning
// like libocr.1.dylib); .so is never used.
func defaultLibraryNamesOS() []string {
	return []string{"libocr.dylib", "libocr.1.dylib", "libocr.0.dylib"}
}
