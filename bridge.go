//go:build cgo

package ocr

/*
#cgo CFLAGS: -I${SRCDIR}
#include <stdlib.h>
#include "ocr.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

// Keep the Go Device constants in sync with the C enum in ocr.h.
const (
	_ = Device(C.OCR_DEVICE_CPU) - DeviceCPU
	_ = Device(C.OCR_DEVICE_GPU) - DeviceGPU
)

// New initializes a native OCR engine with the given configuration.
func New(cfg Config) (*Engine, error) {
	if cfg.ModelDir == "" {
		return nil, errors.New("ocr: Config.ModelDir is required")
	}

	cModelDir := C.CString(cfg.ModelDir)
	defer C.free(unsafe.Pointer(cModelDir))

	var cErr *C.char
	handle := C.ocr_engine_create(
		cModelDir,
		C.int(cfg.Device),
		C.int(cfg.GPUID),
		C.int(cfg.NumThreads),
		&cErr,
	)
	if handle == nil {
		return nil, fmt.Errorf("ocr: engine initialization failed: %w", takeError(cErr))
	}
	return &Engine{handle: unsafe.Pointer(handle)}, nil
}

// Close releases the native engine. It is safe to call more than once.
func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	C.ocr_engine_destroy((*C.OCREngine)(e.handle))
	e.handle = nil
	e.closed = true
	return nil
}

// Recognize runs OCR on an already-encoded image (PNG, JPEG, BMP, ... depending
// on what the backend supports) and returns the detected text blocks.
func (e *Engine) Recognize(image []byte) (*Result, error) {
	if e == nil || e.handle == nil {
		return nil, ErrNotInitialized
	}
	if len(image) == 0 {
		return nil, errors.New("ocr: empty image input")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, errors.New("ocr: engine has been closed")
	}

	var (
		cBlocks *C.OCRBlock
		cCount  C.uint
		cErr    *C.char
	)

	rc := C.ocr_engine_recognize(
		(*C.OCREngine)(e.handle),
		(*C.uchar)(unsafe.Pointer(&image[0])),
		C.ulonglong(len(image)),
		&cBlocks,
		&cCount,
		&cErr,
	)
	if rc != 0 {
		return nil, fmt.Errorf("ocr: recognition failed (code %d): %w", int(rc), takeError(cErr))
	}
	defer C.ocr_free_result(cBlocks)

	n := int(cCount)
	if n == 0 {
		return &Result{Blocks: []Block{}}, nil
	}

	// View the native array as a Go slice without copying it; the backing
	// memory is released by ocr_free_result after conversion below.
	blocks := unsafe.Slice(cBlocks, n)
	out := make([]Block, n)
	for i, b := range blocks {
		out[i] = Block{
			Text:       C.GoString(b.text),
			Confidence: float64(b.confidence),
		}
		for j := 0; j < 4; j++ {
			out[i].Points[j] = Point{
				X: float64(b.xs[j]),
				Y: float64(b.ys[j]),
			}
		}
	}
	return &Result{Blocks: out}, nil
}

// RecognizeFile reads an image file from disk and runs Recognize on it.
func (e *Engine) RecognizeFile(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ocr: read image %q: %w", path, err)
	}
	return e.Recognize(data)
}

// takeError copies a library-owned error message into a Go error and frees
// the native string. A NULL error is replaced with a generic message.
func takeError(cErr *C.char) error {
	if cErr == nil {
		return errors.New("unknown native error (no message provided)")
	}
	msg := C.GoString(cErr)
	C.ocr_free_string(cErr)
	return errors.New(msg)
}

// Version returns the version string reported by the native backend.
func Version() string {
	return C.GoString(C.ocr_version())
}
