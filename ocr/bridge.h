/*
 * bridge.h —— Go 与本地 Tesseract OCR 动态库之间的 C 桥接层。
 *
 * 设计要点：
 *   1. 不 #include <tesseract/capi.h>，这里用不透明指针和与 Tesseract C ABI
 *      完全一致的函数签名自行声明。因此编译机器上无需安装 Tesseract 开发头
 *      文件，只要求运行时能找到对应的动态库。
 *   2. 符号在运行时通过 dlopen / LoadLibrary 解析（见 *_linux.c、*_darwin.c、
 *      *_windows.c），天然屏蔽不同操作系统的库名差异（.so/.dylib/.dll）。
 *   3. 所有函数指针集中在全局 ocr_syms 结构体中，下面的 ocr_* 内联包装函数
 *      统一转发，供 cgo 侧调用。
 */
#ifndef OCR_BRIDGE_H
#define OCR_BRIDGE_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* 与 tesseract/capi.h 中的 TessPageIteratorLevel 保持一致 */
#define OCR_LEVEL_BLOCK    0
#define OCR_LEVEL_PARA     1
#define OCR_LEVEL_TEXTLINE 2
#define OCR_LEVEL_WORD     3
#define OCR_LEVEL_SYMBOL   4

/* ---- 不透明句柄类型（与 libtesseract 内部类型一一对应） ---- */
typedef struct TessBaseAPI        ocr_baseapi_t;
typedef struct TessResultIterator ocr_result_it_t;
typedef struct TessPageIterator   ocr_page_it_t;

/* ---- Tesseract C API 的函数指针原型 ---- */
typedef ocr_baseapi_t *(*ocr_fn_create)(void);
typedef void           (*ocr_fn_delete)(ocr_baseapi_t *);
typedef int            (*ocr_fn_init3)(ocr_baseapi_t *, const char *, const char *);
typedef void           (*ocr_fn_set_psm)(ocr_baseapi_t *, int);
typedef int            (*ocr_fn_recognize)(ocr_baseapi_t *, void *);
typedef void           (*ocr_fn_clear)(ocr_baseapi_t *);
typedef void           (*ocr_fn_end)(ocr_baseapi_t *);
typedef char *         (*ocr_fn_get_utf8)(ocr_baseapi_t *);
typedef int            (*ocr_fn_mean_conf)(ocr_baseapi_t *);
typedef const char *   (*ocr_fn_version)(void);
typedef void           (*ocr_fn_set_image)(ocr_baseapi_t *, const unsigned char *,
                                           int, int, int, int);
typedef ocr_result_it_t *(*ocr_fn_get_iterator)(ocr_baseapi_t *);
typedef void           (*ocr_fn_result_delete)(ocr_result_it_t *);
typedef const char *   (*ocr_fn_ri_text)(ocr_result_it_t *, int);
typedef float          (*ocr_fn_ri_conf)(ocr_result_it_t *, int);
typedef int            (*ocr_fn_ri_next)(ocr_result_it_t *, int);
typedef ocr_page_it_t *(*ocr_fn_ri_get_page)(ocr_result_it_t *);
typedef int            (*ocr_fn_pi_bbox)(ocr_page_it_t *, int,
                                         int *, int *, int *, int *);
typedef void           (*ocr_fn_free_text)(char *);

struct ocr_syms {
    /* 以下符号为必需，缺失任意一个都会导致加载失败 */
    ocr_fn_create        BaseAPICreate;
    ocr_fn_delete        BaseAPIDelete;
    ocr_fn_init3         BaseAPIInit3;
    ocr_fn_set_psm       BaseAPISetPageSegMode;
    ocr_fn_recognize     BaseAPIRecognize;
    ocr_fn_clear         BaseAPIClear;
    ocr_fn_end           BaseAPIEnd;
    ocr_fn_get_utf8      BaseAPIGetUTF8Text;
    ocr_fn_mean_conf     BaseAPIMeanTextConf;
    ocr_fn_set_image     BaseAPISetImage;
    ocr_fn_get_iterator  BaseAPIGetIterator;
    ocr_fn_result_delete ResultIteratorDelete;
    ocr_fn_ri_text       ResultIteratorGetUTF8Text;
    ocr_fn_ri_conf       ResultIteratorConfidence;
    ocr_fn_ri_next       ResultIteratorNext;
    ocr_fn_ri_get_page   ResultIteratorGetPageIterator;
    ocr_fn_pi_bbox       PageIteratorBoundingBox;
    ocr_fn_free_text     DeleteText;
    /* 可选符号，缺失不影响加载 */
    ocr_fn_version       Version;
};

/* 各平台 *.c 文件中唯一定义的全局符号表与句柄 */
extern struct ocr_syms ocr_syms;

/*
 * 加载动态库并解析全部必需符号。
 *   override_lib：显式指定的库路径，为空则查 OCR_LIBRARY_PATH 环境变量，
 *                 再退化为当前操作系统的默认候选列表。
 *   err/errlen  ：失败时写入人类可读的错误信息。
 * 返回 0 成功，非 0 失败。
 */
int  ocr_load_library(const char *override_lib, char *err, size_t errlen);
void ocr_unload_library(void);

/* ---- 对 Go 暴露的薄包装：转发到函数指针 ---- */
static inline ocr_baseapi_t *ocr_baseapi_create(void) {
    return ocr_syms.BaseAPICreate();
}
static inline void ocr_baseapi_delete(ocr_baseapi_t *a) {
    ocr_syms.BaseAPIDelete(a);
}
static inline int ocr_baseapi_init3(ocr_baseapi_t *a, const char *d, const char *l) {
    return ocr_syms.BaseAPIInit3(a, d, l);
}
static inline void ocr_baseapi_set_psm(ocr_baseapi_t *a, int m) {
    ocr_syms.BaseAPISetPageSegMode(a, m);
}
static inline int ocr_baseapi_recognize(ocr_baseapi_t *a) {
    return ocr_syms.BaseAPIRecognize(a, NULL);
}
static inline void ocr_baseapi_clear(ocr_baseapi_t *a) {
    ocr_syms.BaseAPIClear(a);
}
static inline void ocr_baseapi_end(ocr_baseapi_t *a) {
    ocr_syms.BaseAPIEnd(a);
}
static inline char *ocr_baseapi_get_utf8(ocr_baseapi_t *a) {
    return ocr_syms.BaseAPIGetUTF8Text(a);
}
static inline int ocr_baseapi_mean_conf(ocr_baseapi_t *a) {
    return ocr_syms.BaseAPIMeanTextConf(a);
}
static inline const char *ocr_version(void) {
    return ocr_syms.Version ? ocr_syms.Version() : NULL;
}
static inline void ocr_baseapi_set_image(ocr_baseapi_t *a,
                                         const unsigned char *pix,
                                         int w, int h, int stride) {
    ocr_syms.BaseAPISetImage(a, pix, w, h, 4, stride);
}
static inline ocr_result_it_t *ocr_baseapi_get_iterator(ocr_baseapi_t *a) {
    return ocr_syms.BaseAPIGetIterator(a);
}
static inline void ocr_result_delete(ocr_result_it_t *it) {
    ocr_syms.ResultIteratorDelete(it);
}
static inline const char *ocr_result_text(ocr_result_it_t *it, int level) {
    return ocr_syms.ResultIteratorGetUTF8Text(it, level);
}
static inline float ocr_result_conf(ocr_result_it_t *it, int level) {
    return ocr_syms.ResultIteratorConfidence(it, level);
}
static inline int ocr_result_next(ocr_result_it_t *it, int level) {
    return ocr_syms.ResultIteratorNext(it, level);
}
static inline int ocr_result_bbox(ocr_result_it_t *it, int level,
                                  int *l, int *t, int *r, int *b) {
    /* BoundingBox 挂在 PageIterator 上；由 ResultIterator 取得，
     * 其生命周期依附于 ResultIterator，无需单独释放。 */
    ocr_page_it_t *pi = ocr_syms.ResultIteratorGetPageIterator(it);
    if (pi == NULL) {
        return 0;
    }
    return ocr_syms.PageIteratorBoundingBox(pi, level, l, t, r, b);
}
static inline void ocr_free_text(char *p) {
    ocr_syms.DeleteText(p);
}

#ifdef __cplusplus
}
#endif

#endif /* OCR_BRIDGE_H */
