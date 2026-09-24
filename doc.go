// Package ocr binds a local OCR shared library (libocr.so / libocr.dylib /
// ocr.dll) through cgo and exposes a small, memory-safe Go API.
//
// The native side must export the C ABI declared in ocr.h: create/destroy an
// engine, recognize an encoded image into text blocks with confidence, and
// free the returned results.
//
// Quick start:
//
//	eng, err := ocr.New(ocr.Config{ModelDir: "./models"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer eng.Close()
//
//	res, err := eng.RecognizeFile("invoice.png")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, b := range res.Blocks {
//	    fmt.Printf("%s (%.2f)\n", b.Text, b.Confidence)
//	}
//
// Platform library names and link flags are handled in link_*.go and
// libname_*.go; see README.md for deployment details.
package ocr
