//go:build cgo

package ocr

import (
	"os"
	"path/filepath"
	"testing"
)

// These tests call into the native ABI. They are skipped unless an OCR shared
// library is reachable: either a real backend installed system-wide, or the
// reference stub under test/stub (build it with `make stub`), exposed via
// LD_LIBRARY_PATH / DYLD_LIBRARY_PATH / PATH.
//
// To force the suite to fail when the library is missing instead of skipping,
// set GOOCR_REQUIRE_LIB=1.

func requireLibrary(t *testing.T) {
	t.Helper()
	if _, err := FindLibrary(); err != nil {
		if os.Getenv("GOOCR_REQUIRE_LIB") == "1" {
			t.Fatalf("native OCR library required but not found: %v", err)
		}
		t.Skipf("native OCR library not available: %v", err)
	}
}

func TestVersion(t *testing.T) {
	requireLibrary(t)
	if v := Version(); v == "" {
		t.Fatal("Version() returned empty string")
	}
}

func TestNewRejectsEmptyModelDir(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New with empty ModelDir should fail")
	}
}

func TestRecognizeEmptyImage(t *testing.T) {
	requireLibrary(t)
	eng, err := New(Config{ModelDir: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })

	if _, err := eng.Recognize(nil); err == nil {
		t.Fatal("Recognize(nil) should fail")
	}
}

func TestRecognizeStubRoundTrip(t *testing.T) {
	requireLibrary(t)
	eng, err := New(Config{ModelDir: t.TempDir(), NumThreads: 2})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := eng.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Double close must be harmless.
	if err := eng.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestRecognizeFileWithStub(t *testing.T) {
	requireLibrary(t)

	modelDir := t.TempDir()
	eng, err := New(Config{ModelDir: modelDir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })

	img := filepath.Join(t.TempDir(), "img.png")
	if err := os.WriteFile(img, []byte("Pretend this is a PNG payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := eng.RecognizeFile(img)
	if err != nil {
		t.Fatalf("RecognizeFile: %v", err)
	}
	if len(res.Blocks) == 0 {
		t.Fatal("expected at least one block from the native backend")
	}
	for i, b := range res.Blocks {
		if b.Text == "" {
			t.Errorf("block %d has empty text", i)
		}
		if b.Confidence < 0 || b.Confidence > 1 {
			t.Errorf("block %d confidence out of range: %v", i, b.Confidence)
		}
		for j, p := range b.Points {
			if p.X < 0 || p.Y < 0 {
				t.Errorf("block %d point %d negative: %+v", i, j, p)
			}
		}
	}

	if txt := res.FullText(); txt == "" {
		t.Error("FullText() is empty")
	}
}

func TestRecognizeMissingFile(t *testing.T) {
	requireLibrary(t)
	eng, err := New(Config{ModelDir: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })

	if _, err := eng.RecognizeFile(filepath.Join(t.TempDir(), "nope.png")); err == nil {
		t.Fatal("RecognizeFile on missing path should fail")
	}
}
