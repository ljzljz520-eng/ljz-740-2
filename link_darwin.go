//go:build cgo && darwin && !ocr_custom_lib

package ocr

/*
#cgo LDFLAGS: -locr
*/
import "C"
