package ocr

import (
	"errors"
	"sync"
	"unsafe"
)

// Device selects the compute device used by the OCR backend.
//
// The numeric values mirror the C enum in ocr.h; they are duplicated here so
// that the public types remain available when the package is built without
// cgo (see stub_nocgo.go).
type Device int

const (
	DeviceCPU Device = 0
	DeviceGPU Device = 1
)

// Point is one corner of a detected text quadrilateral, in pixels.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Block is one recognized text region: the UTF-8 text, its confidence
// (0.0-1.0) and the four corner points of the text quadrilateral, ordered
// clockwise starting from the top-left corner.
type Block struct {
	Text       string   `json:"text"`
	Confidence float64  `json:"confidence"`
	Points     [4]Point `json:"points"`
}

// Result is the output of a recognition call.
type Result struct {
	Blocks []Block `json:"blocks"`
}

// FullText joins all blocks with new lines.
func (r *Result) FullText() string {
	if r == nil || len(r.Blocks) == 0 {
		return ""
	}
	out := make([]byte, 0, 256)
	for i, b := range r.Blocks {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, b.Text...)
	}
	return string(out)
}

// Config configures an Engine.
type Config struct {
	// ModelDir is the directory containing the OCR model files. Required
	// by most backends; the contents are backend-specific.
	ModelDir string
	// Device selects CPU or GPU inference. Defaults to CPU.
	Device Device
	// GPUID selects the GPU when Device == DeviceGPU. Ignored on CPU.
	GPUID int
	// NumThreads sets the number of CPU inference threads. Zero lets the
	// backend choose its own default.
	NumThreads int
}

// Engine is a handle to a native OCR engine instance.
//
// A C engine is typically backed by one inference session and must not be
// driven concurrently, so the cgo method set serializes calls through mu.
// An Engine must not be copied after creation.
type Engine struct {
	handle unsafe.Pointer
	mu     sync.Mutex
	closed bool
}

// ErrNotInitialized is returned when an Engine method is called before New
// succeeded or after Close.
var ErrNotInitialized = errors.New("ocr: engine is not initialized")
