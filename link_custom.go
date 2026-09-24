//go:build cgo && ocr_custom_lib

// When the ocr_custom_lib build tag is set no -locr flag is emitted; supply
// the native library yourself through CGO_LDFLAGS, for example:
//
//	CGO_LDFLAGS="-L/opt/ocr/lib -l:libocr.so.1" go build -tags ocr_custom_lib ./...
package ocr
