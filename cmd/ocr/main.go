// Command ocr recognizes text in an image using the local OCR engine binding
// and prints the result (text blocks + confidence) as JSON.
//
// Example:
//
//	ocr -model ./models -image invoice.png
//	ocr -model ./models -image scan.jpg -device gpu -gpu-id 0 -pretty
//	ocr -version
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	ocr "github.com/solo-manager/goocr"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ocr:", err)
		os.Exit(1)
	}
}

// jsonOutput is the stable CLI payload printed on stdout.
type jsonOutput struct {
	Backend string      `json:"backend,omitempty"`
	Image   string      `json:"image,omitempty"`
	Count   int         `json:"count"`
	Result  *ocr.Result `json:"result"`
}

func run(args []string) error {
	fs := flag.NewFlagSet("ocr", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "ocr: recognize an image with the local OCR engine and print JSON.\n")
		fmt.Fprintf(fs.Output(), "\nThe native shared library (libocr.so / libocr.dylib / ocr.dll) is loaded\n")
		fmt.Fprintf(fs.Output(), "at process start; put it on the loader path (LD_LIBRARY_PATH /\n")
		fmt.Fprintf(fs.Output(), "DYLD_LIBRARY_PATH / PATH) or in a system library directory.\n\n")
		fmt.Fprintf(fs.Output(), "Usage: ocr -model <dir> -image <file> [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}
	var (
		modelDir  = fs.String("model", "", "path to the OCR model directory (required)")
		imagePath = fs.String("image", "", "path to the image file to recognize (required)")
		device    = fs.String("device", "cpu", "compute device: cpu or gpu")
		gpuID     = fs.Int("gpu-id", 0, "GPU id used when -device=gpu")
		threads   = fs.Int("threads", 0, "number of CPU inference threads (0 = backend default)")
		pretty    = fs.Bool("pretty", false, "pretty-print the JSON output")
		showVer   = fs.Bool("version", false, "print native backend version and exit")
		checkLib  = fs.Bool("check-library", false, "locate the native shared library and exit")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVer {
		fmt.Println(ocr.Version())
		return nil
	}

	if *checkLib {
		path, err := ocr.FindLibrary()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	}

	if *modelDir == "" {
		return errors.New("missing -model: path to the model directory is required")
	}
	if *imagePath == "" {
		return errors.New("missing -image: path to an image file is required")
	}

	dev := ocr.DeviceCPU
	switch *device {
	case "cpu", "CPU":
		dev = ocr.DeviceCPU
	case "gpu", "GPU":
		dev = ocr.DeviceGPU
	default:
		return fmt.Errorf("invalid -device %q: want cpu or gpu", *device)
	}

	engine, err := ocr.New(ocr.Config{
		ModelDir:   *modelDir,
		Device:     dev,
		GPUID:      *gpuID,
		NumThreads: *threads,
	})
	if err != nil {
		return err
	}
	defer engine.Close()

	result, err := engine.RecognizeFile(*imagePath)
	if err != nil {
		return err
	}

	out := jsonOutput{
		Backend: ocr.Version(),
		Image:   *imagePath,
		Count:   len(result.Blocks),
		Result:  result,
	}

	enc := json.NewEncoder(os.Stdout)
	if *pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	return nil
}
