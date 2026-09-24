/*
 * lib_windows.c —— Windows 平台动态库加载（仅在 GOOS=windows 时参与编译，
 * 需要使用 MinGW 工具链进行 cgo 编译）。
 *
 * Windows 下 Tesseract 的动态库为 tesseract.dll。官方 Windows 安装包
 * （UB-Mannheim 构建）通常把它放在安装根目录或 bin 目录下，因此除裸 DLL
 * 名（走标准搜索顺序：程序目录 / PATH）外，这里还列出几个最常见的安装路径。
 */
#include <windows.h>
#include <stdio.h>
#include <string.h>

#include "bridge.h"

struct ocr_syms ocr_syms;
static HMODULE ocr_handle = NULL;

/* 必需符号：解析失败直接置错误并返回 -1 */
#define OCR_REQUIRE(field, symbol)                                            \
    do {                                                                      \
        ocr_syms.field = (__typeof__(ocr_syms.field))GetProcAddress(            \
            ocr_handle, symbol);                                                 \
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

    /* 优先级：显式参数 > OCR_LIBRARY_PATH 环境变量 > 默认候选 */
    if (override_lib != NULL && override_lib[0] != '\0') {
        candidates[n++] = override_lib;
    } else {
        env = getenv("OCR_LIBRARY_PATH");
        if (env != NULL && env[0] != '\0') {
            candidates[n++] = env;
        }
    }
    if (n == 0) {
        const char *defaults[] = {
            "tesseract54.dll",
            "tesseract53.dll",
            "tesseract51.dll",
            "tesseract50.dll",
            "tesseract41.dll",
            "tesseract.dll",
            "C:\\Program Files\\Tesseract-OCR\\tesseract54.dll",
            "C:\\Program Files\\Tesseract-OCR\\tesseract53.dll",
            "C:\\Program Files\\Tesseract-OCR\\tesseract.dll",
            "C:\\Program Files (x86)\\Tesseract-OCR\\tesseract.dll",
        };
        size_t i;
        for (i = 0; i < sizeof(defaults) / sizeof(defaults[0]) && n < 64; i++) {
            candidates[n++] = defaults[i];
        }
    }

    size_t i;
    for (i = 0; i < n; i++) {
        ocr_handle = LoadLibraryA(candidates[i]);
        if (ocr_handle != NULL) {
            break;
        }
    }

    if (ocr_handle == NULL) {
        DWORD code = GetLastError();
        snprintf(err, errlen,
                 "无法加载 OCR 动态库（已尝试 %u 个候选路径）。"
                 "GetLastError=%lu；请确认 tesseract.dll 位于 PATH 中，"
                 "或通过 OCR_LIBRARY_PATH 指定完整路径。",
                 (unsigned)n, (unsigned long)code);
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

    ocr_syms.Version = (__typeof__(ocr_syms.Version))GetProcAddress(
        ocr_handle, "TessVersion");

    return 0;
}

void ocr_unload_library(void) {
    if (ocr_handle != NULL) {
        FreeLibrary(ocr_handle);
        ocr_handle = NULL;
        memset(&ocr_syms, 0, sizeof(ocr_syms));
    }
}
