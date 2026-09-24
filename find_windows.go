//go:build windows

package ocr

import "os"

func systemLibraryDirs() []string {
	// On Windows the loader searches System32 automatically; include it
	// explicitly so FindLibrary reports a path as well.
	dirs := []string{`C:\Windows\System32`, `C:\Windows`}
	if windir := os.Getenv("WINDIR"); windir != "" {
		dirs[0] = windir + `\System32`
		dirs[1] = windir
	}
	return dirs
}
