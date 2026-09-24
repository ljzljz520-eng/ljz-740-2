//go:build cgo && windows && !ocr_custom_lib

/*
On Windows -locr expects an import library named ocr.lib (MSVC) or
libocr.dll.a / ocr.dll.a (MinGW-w64). If you only have ocr.dll you can link
against it directly, e.g.:

	go build -tags ocr_custom_lib .
	CGO_LDFLAGS="-L/path/to/dll -l:ocr.dll" go build .
*/
package ocr

/*
#cgo LDFLAGS: -locr
*/
import "C"
