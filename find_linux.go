//go:build linux

package ocr

func systemLibraryDirs() []string {
	return []string{
		"/usr/local/lib",
		"/usr/lib",
		"/usr/lib64",
		"/lib",
		"/lib64",
		// Common multiarch tuples (Debian/Ubuntu, Fedora/RHEL).
		"/usr/lib/x86_64-linux-gnu",
		"/usr/lib/aarch64-linux-gnu",
		"/usr/lib64/x86_64-linux-gnu",
	}
}
