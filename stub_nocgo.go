//go:build !cgo

// This file keeps the package API intact when built with CGO_ENABLED=0.
//
// The native OCR library can only be driven through cgo, so every function
// that would cross the FFI boundary returns an explicit error telling the
// caller to rebuild with cgo enabled and the shared library present.
package ocr

import "errors"

var errCGODisabled = errors.New("ocr: package built with CGO_ENABLED=0; rebuild with cgo enabled and the native OCR library installed")

// New always fails when cgo is disabled.
func New(Config) (*Engine, error) { return nil, errCGODisabled }

// Close is a no-op stub when cgo is disabled.
func (e *Engine) Close() error { return nil }

// Recognize always fails when cgo is disabled.
func (e *Engine) Recognize([]byte) (*Result, error) { return nil, errCGODisabled }

// RecognizeFile always fails when cgo is disabled.
func (e *Engine) RecognizeFile(string) (*Result, error) { return nil, errCGODisabled }

// Version reports that the native binding is unavailable.
func Version() string { return "unavailable (CGO_ENABLED=0)" }
