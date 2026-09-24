package ocr

// defaultLibraryNames returns candidate base names of the native OCR shared
// library for the current operating system. The first entry is the preferred
// name; FindLibrary probes the remaining fallbacks in order.
//
// Per-OS overrides live in libname_*.go.
func defaultLibraryNames() []string {
	return defaultLibraryNamesOS()
}
