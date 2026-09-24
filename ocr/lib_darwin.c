/*
 * lib_darwin.c —— macOS 平台动态库候选名（仅在 GOOS=darwin 时参与编译）。
 *
 * macOS 下 Tesseract 的动态库为 libtesseract.dylib，常见位置：
 *   - Homebrew(Apple Silicon): /opt/homebrew/lib
 *   - Homebrew(Intel):         /usr/local/lib
 *   - MacPorts:                /opt/local/lib
 * 版本化文件名形如 libtesseract.5.dylib（注意与 Linux 的 .so.5 顺序不同）。
 */

#define OCR_DEFAULT_LIBS {                                    \
    "libtesseract.5.dylib",                                   \
    "libtesseract.4.dylib",                                   \
    "libtesseract.dylib",                                     \
    "/opt/homebrew/lib/libtesseract.5.dylib",                 \
    "/opt/homebrew/lib/libtesseract.dylib",                   \
    "/usr/local/lib/libtesseract.5.dylib",                    \
    "/usr/local/lib/libtesseract.dylib",                      \
    "/opt/local/lib/libtesseract.5.dylib",                    \
    "/opt/local/lib/libtesseract.dylib",                      \
}

#include "loader_posix.h"
