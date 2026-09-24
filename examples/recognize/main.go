// Minimal library-API example: recognize an image file and print JSON blocks.
//
// Run with the native library reachable by the loader, e.g. with the stub:
//
//	make stub
//	CGO_LDFLAGS="-L$(pwd)/test/stub" \
//	LD_LIBRARY_PATH="$(pwd)/test/stub" \
//	go run ./examples/recognize /path/to/image.png
package main

import (
	"encoding/json"
	"fmt"
	"os"

	ocr "github.com/solo-manager/goocr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: recognize <image>")
		os.Exit(2)
	}
	modelDir := os.Getenv("OCR_MODEL_DIR")
	if modelDir == "" {
		modelDir = "./models"
	}

	eng, err := ocr.New(ocr.Config{ModelDir: modelDir, Device: ocr.DeviceCPU})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer eng.Close()

	res, err := eng.RecognizeFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(struct {
		Backend string      `json:"backend"`
		Result  *ocr.Result `json:"result"`
	}{
		Backend: ocr.Version(),
		Result:  res,
	}, "", "  ")
	fmt.Println(string(out))
}
