/*
 * ocr_stub.c - minimal reference implementation of the ocr.h ABI.
 *
 * It performs no real OCR: engine creation only validates the model path,
 * and recognition returns deterministic fake blocks. Its purpose is to let
 * the Go binding and the CLI be built and exercised end-to-end without a
 * multi-hundred-MB model / inference backend.
 *
 * Build:
 *   Linux:   cc -shared -fPIC -I../.. -DOCR_BUILDING_LIBRARY \
 *              -o libocr.so ocr_stub.c
 *   macOS:   cc -dynamiclib -I../.. -DOCR_BUILDING_LIBRARY \
 *              -o libocr.dylib ocr_stub.c
 *   Windows: see test/stub/Makefile
 */
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "ocr.h"

struct OCREngine {
    int device;
    int gpu_id;
    int num_threads;
};

static char* dup_error(const char* msg) {
    size_t n = strlen(msg) + 1;
    char* p = (char*)malloc(n);
    if (p) memcpy(p, msg, n);
    return p;
}

OCR_API OCREngine* ocr_engine_create(const char* model_dir,
                                     int device,
                                     int gpu_id,
                                     int num_threads,
                                     char** err) {
    if (err) *err = NULL;
    if (!model_dir || !*model_dir) {
        if (err) *err = dup_error("model_dir must not be empty");
        return NULL;
    }
    if (device != OCR_DEVICE_CPU && device != OCR_DEVICE_GPU) {
        if (err) *err = dup_error("unsupported device type");
        return NULL;
    }
    OCREngine* e = (OCREngine*)calloc(1, sizeof(OCREngine));
    if (!e) {
        if (err) *err = dup_error("out of memory");
        return NULL;
    }
    e->device = device;
    e->gpu_id = gpu_id;
    e->num_threads = num_threads;
    return e;
}

OCR_API void ocr_engine_destroy(OCREngine* engine) {
    free(engine);
}

OCR_API int ocr_engine_recognize(OCREngine* engine,
                                 const unsigned char* image,
                                 unsigned long long image_len,
                                 OCRBlock** out_blocks,
                                 unsigned int* out_count,
                                 char** err) {
    if (err) *err = NULL;
    *out_blocks = NULL;
    *out_count = 0;

    if (!engine) {
        if (err) *err = dup_error("nil engine");
        return 1;
    }
    if (!image || image_len == 0) {
        if (err) *err = dup_error("empty image");
        return 2;
    }

    OCRBlock* blocks = (OCRBlock*)calloc(2, sizeof(OCRBlock));
    if (!blocks) {
        if (err) *err = dup_error("out of memory");
        return 3;
    }

    blocks[0].text = strdup("Hello, OCR!");
    blocks[0].confidence = 0.97;
    blocks[0].xs[0] = 12;  blocks[0].ys[0] = 8;
    blocks[0].xs[1] = 202; blocks[0].ys[1] = 8;
    blocks[0].xs[2] = 202; blocks[0].ys[2] = 40;
    blocks[0].xs[3] = 12;  blocks[0].ys[3] = 40;

    blocks[1].text = strdup("stub backend (model: stub)");
    blocks[1].confidence = 0.88;
    blocks[1].xs[0] = 12;  blocks[1].ys[0] = 52;
    blocks[1].xs[1] = 360; blocks[1].ys[1] = 52;
    blocks[1].xs[2] = 360; blocks[1].ys[2] = 84;
    blocks[1].xs[3] = 12;  blocks[1].ys[3] = 84;

    *out_blocks = blocks;
    *out_count = 2;
    return 0;
}

OCR_API void ocr_free_result(OCRBlock* blocks) {
    /* The stub strdup()s each text, so free them individually. */
    if (blocks) {
        free((void*)blocks[0].text);
        free((void*)blocks[1].text);
        free(blocks);
    }
}

OCR_API const char* ocr_version(void) {
    return "stub-ocr/1.0.0";
}

OCR_API void ocr_free_string(char* s) {
    free(s);
}
