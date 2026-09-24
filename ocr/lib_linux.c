/*
 * lib_linux.c —— Linux 平台动态库候选名（仅在 GOOS=linux 时参与编译）。
 *
 * Linux 下 Tesseract 的动态库为 libtesseract.so，常见形态：
 *   - 开发包安装的带版本号 .so 软链：libtesseract.so
 *   - 运行时真实文件：libtesseract.so.5 / libtesseract.so.4 / libtesseract.so.3
 * dlopen 对裸 soname 会走 ldconfig 缓存，因此多数系统只需 "libtesseract.so"。
 */

/* Linux 库名候选：从最通用的 soname 到各主版本 */
#define OCR_DEFAULT_LIBS {                       \
    "libtesseract.so.5",                         \
    "libtesseract.so.4",                         \
    "libtesseract.so.3",                         \
    "libtesseract.so",                           \
    "/usr/lib/x86_64-linux-gnu/libtesseract.so.5",  \
    "/usr/lib/x86_64-linux-gnu/libtesseract.so.4",  \
    "/usr/lib/libtesseract.so.5",                \
    "/usr/lib64/libtesseract.so.5",              \
    "/usr/local/lib/libtesseract.so.5",          \
    "/usr/local/lib/libtesseract.so",            \
}

#include "loader_posix.h"
