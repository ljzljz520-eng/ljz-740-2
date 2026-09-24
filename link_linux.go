//go:build cgo && linux && !ocr_custom_lib

package ocr

/*
#cgo LDFLAGS: -locr
*/
import "C"
