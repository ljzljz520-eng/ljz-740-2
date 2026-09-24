/*
 * ocr.h - C ABI for the local OCR engine shared library.
 *
 * The Go package in this directory binds to a shared library that exports
 * the functions declared below. Concrete OCR backends (PaddleOCR, Tesseract,
 * RapidOCR, a vendor SDK, ...) only need to implement this small stable ABI.
 *
 * Platform library file names:
 *   Linux:   libocr.so      (also accepts versioned names, e.g. libocr.so.1)
 *   macOS:   libocr.dylib
 *   Windows: ocr.dll
 */
#ifndef GO_OCR_ENGINE_H
#define GO_OCR_ENGINE_H

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN32) || defined(__CYGWIN__)
  #ifdef OCR_BUILDING_LIBRARY
    #define OCR_API __declspec(dllexport)
  #else
    #define OCR_API __declspec(dllimport)
  #endif
#else
  #define OCR_API
#endif

/* Opaque engine handle returned by ocr_engine_create. */
typedef struct OCREngine OCREngine;

/* Compute device to run inference on. */
typedef enum {
    OCR_DEVICE_CPU = 0,
    OCR_DEVICE_GPU = 1
} OCRDevice;

/*
 * One detected text line / block.
 *
 * Coordinates are stored as 4 corner points of the (possibly rotated)
 * quadrilateral in clockwise order starting at the top-left corner:
 * xs[0],ys[0] ... xs[3],ys[3].
 * All pointers are valid until the next call on the same engine or until
 * the result is freed with ocr_free_result.
 */
typedef struct {
    const char* text;       /* UTF-8 encoded text */
    double      confidence; /* 0.0 .. 1.0 */
    double      xs[4];
    double      ys[4];
} OCRBlock;

/* Create an engine instance. Returns NULL and fills *err on failure. */
OCR_API OCREngine* ocr_engine_create(const char* model_dir,
                                     int device,
                                     int gpu_id,
                                     int num_threads,
                                     char** err);

/* Free an engine instance. Safe to call with NULL. */
OCR_API void ocr_engine_destroy(OCREngine* engine);

/*
 * Recognize text in an encoded image (PNG/JPEG/BMP/... bytes).
 * On success returns 0 and fills *out_blocks with a heap array of *out_count
 * OCRBlock items owned by the library; release with ocr_free_result.
 * On failure returns a non-zero code and fills *err.
 */
OCR_API int ocr_engine_recognize(OCREngine* engine,
                                 const unsigned char* image,
                                 unsigned long long image_len,
                                 OCRBlock** out_blocks,
                                 unsigned int* out_count,
                                 char** err);

/* Free the block array produced by ocr_engine_recognize. */
OCR_API void ocr_free_result(OCRBlock* blocks);

/* Version string of the underlying backend, e.g. "paddleocr-2.7.0". */
OCR_API const char* ocr_version(void);

/* Free an error string previously returned by the library. */
OCR_API void ocr_free_string(char* s);

#ifdef __cplusplus
}
#endif

#endif /* GO_OCR_ENGINE_H */
