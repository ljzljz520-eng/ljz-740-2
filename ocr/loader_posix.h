/*
 * loader_posix.h —— POSIX（Linux / macOS）共享的动态库加载逻辑。
 * 各平台 .c 文件只需定义 OCR_DEFAULT_LIBS（默认库候选名数组）后引入本文件。
 */
#ifndef OCR_LOADER_POSIX_H
#define OCR_LOADER_POSIX_H

#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "bridge.h"

struct ocr_syms ocr_syms;
static void *ocr_handle = NULL;

#ifndef OCR_DEFAULT_LIBS
#error "请在平台 .c 文件中先定义 OCR_DEFAULT_LIBS"
#endif

static void *ocr_resolve(const char *name) {
    return dlsym(ocr_handle, name);
}

/* 必需符号：解析失败直接置错误并返回 -1 */
#define OCR_REQUIRE(field, symbol)                                            \
    do {                                                                      \
        *(void **)(&ocr_syms.field) = ocr_resolve(symbol);                    \
        if (ocr_syms.field == NULL) {                                         \
            snprintf(err, errlen, "动态库中缺少必需符号: %s", symbol);        \
            return -1;                                                        \
        }                                                                     \
    } while (0)

int ocr_load_library(const char *override_lib, char *err, size_t errlen) {
    const char *env = NULL;
    const char *candidates[64];
    size_t n = 0;

    if (ocr_handle != NULL) {
        return 0;
    }

    /* 优先级：显式参数 > OCR_LIBRARY_PATH 环境变量 > 平台默认候选 */
    if (override_lib != NULL && override_lib[0] != '\0') {
        candidates[n++] = override_lib;
    } else {
        env = getenv("OCR_LIBRARY_PATH");
        if (env != NULL && env[0] != '\0') {
            candidates[n++] = env;
        }
    }

    const char *last_error = "(无)";
    if (n == 0) {
        const char *defaults[] = OCR_DEFAULT_LIBS;
        size_t i;
        for (i = 0; i < sizeof(defaults) / sizeof(defaults[0]) && n < 64; i++) {
            candidates[n++] = defaults[i];
        }
    }

    size_t i;
    for (i = 0; i < n; i++) {
        dlerror(); /* 清空旧错误 */
        ocr_handle = dlopen(candidates[i], RTLD_NOW | RTLD_LOCAL);
        if (ocr_handle != NULL) {
            break;
        }
        const char *e = dlerror();
        if (e != NULL) {
            last_error = e;
        }
    }

    if (ocr_handle == NULL) {
        snprintf(err, errlen,
                 "无法加载 OCR 动态库（已尝试 %zu 个候选路径）。最后一次错误: %s",
                 n, last_error);
        return -1;
    }

    memset(&ocr_syms, 0, sizeof(ocr_syms));

    OCR_REQUIRE(BaseAPICreate,                "TessBaseAPICreate");
    OCR_REQUIRE(BaseAPIDelete,                "TessBaseAPIDelete");
    OCR_REQUIRE(BaseAPIInit3,                 "TessBaseAPIInit3");
    OCR_REQUIRE(BaseAPISetPageSegMode,        "TessBaseAPISetPageSegMode");
    OCR_REQUIRE(BaseAPIRecognize,             "TessBaseAPIRecognize");
    OCR_REQUIRE(BaseAPIClear,                 "TessBaseAPIClear");
    OCR_REQUIRE(BaseAPIEnd,                   "TessBaseAPIEnd");
    OCR_REQUIRE(BaseAPIGetUTF8Text,           "TessBaseAPIGetUTF8Text");
    OCR_REQUIRE(BaseAPIMeanTextConf,          "TessBaseAPIMeanTextConf");
    OCR_REQUIRE(BaseAPISetImage,              "TessBaseAPISetImage");
    OCR_REQUIRE(BaseAPIGetIterator,           "TessBaseAPIGetIterator");
    OCR_REQUIRE(ResultIteratorDelete,         "TessResultIteratorDelete");
    OCR_REQUIRE(ResultIteratorGetUTF8Text,    "TessResultIteratorGetUTF8Text");
    OCR_REQUIRE(ResultIteratorConfidence,     "TessResultIteratorConfidence");
    OCR_REQUIRE(ResultIteratorNext,           "TessResultIteratorNext");
    OCR_REQUIRE(ResultIteratorGetPageIterator,"TessResultIteratorGetPageIterator");
    OCR_REQUIRE(PageIteratorBoundingBox,      "TessPageIteratorBoundingBox");
    OCR_REQUIRE(DeleteText,                   "TessDeleteText");

    /* 可选符号 */
    *(void **)(&ocr_syms.Version) = ocr_resolve("TessVersion");

    return 0;
}

void ocr_unload_library(void) {
    if (ocr_handle != NULL) {
        dlclose(ocr_handle);
        ocr_handle = NULL;
        memset(&ocr_syms, 0, sizeof(ocr_syms));
    }
}

#endif /* OCR_LOADER_POSIX_H */
