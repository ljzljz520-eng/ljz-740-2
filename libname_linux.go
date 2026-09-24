//go:build linux

package ocr

// On Linux the ELF shared object is conventionally called libocr.so;
// distros often ship a versioned runtime name such as libocr.so.1.
func defaultLibraryNamesOS() []string {
	return []string{"libocr.so", "libocr.so.1", "libocr.so.0"}
}
