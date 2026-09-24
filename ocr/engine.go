package ocr

/*
#cgo CFLAGS: -std=c11 -Wno-unused-function
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"sync"
	"unsafe"
)

// 动态库在进程内只加载一次；firstError 记录第一次加载尝试的结果。
var (
	loadMu     sync.Mutex
	loadLoaded bool
	loadError  error
)

// ensureLoaded 确保本地 OCR 动态库已加载。
// 首次调用时使用的 overridePath 决定加载哪个库。
func ensureLoaded(overridePath string) error {
	loadMu.Lock()
	defer loadMu.Unlock()

	if loadLoaded {
		return loadError
	}

	var cPath *C.char
	if overridePath != "" {
		cPath = C.CString(overridePath)
		defer C.free(unsafe.Pointer(cPath))
	}

	var cErr [512]C.char
	rc := C.ocr_load_library(cPath, &cErr[0], C.size_t(len(cErr)))
	if rc != 0 {
		loadError = errors.New(C.GoString(&cErr[0]))
		return loadError
	}

	loadLoaded = true
	return nil
}

// Engine 是一个 Tesseract OCR 引擎实例。
//
// Tesseract 单实例不可并发使用，因此 Engine 内部持有互斥锁；多个 goroutine
// 可共享同一 *Engine，但识别会串行执行。需要真正并行时请创建多个 Engine。
type Engine struct {
	mu     sync.Mutex
	api    unsafe.Pointer // *TessBaseAPI（不透明）
	level  Level
	closed bool
}

// New 加载本地动态库并初始化一个 OCR 引擎实例（加载模型/语言数据）。
// cfg 为零值时等价于默认英文模型、按文本块返回结果。
func New(cfg Config) (*Engine, error) {
	lang := cfg.Language
	if lang == "" {
		lang = "eng"
	}
	level := cfg.Level

	if err := ensureLoaded(cfg.LibraryPath); err != nil {
		return nil, fmt.Errorf("ocr: 加载本地动态库失败: %w", err)
	}

	api := unsafe.Pointer(C.ocr_baseapi_create())
	if api == nil {
		return nil, errors.New("ocr: TessBaseAPICreate 返回空句柄")
	}

	cData := C.CString(cfg.DataPath)
	defer C.free(unsafe.Pointer(cData))
	cLang := C.CString(lang)
	defer C.free(unsafe.Pointer(cLang))

	// Init3 返回 0 表示成功；失败通常是找不到 tessdata 目录或语言训练文件。
	if rc := C.ocr_baseapi_init3((*C.ocr_baseapi_t)(api), cData, cLang); rc != 0 {
		C.ocr_baseapi_delete((*C.ocr_baseapi_t)(api))
		return nil, fmt.Errorf(
			"ocr: 初始化模型失败（语言 %q，tessdata 路径 %q）；"+
				"请确认对应 .traineddata 已安装，或通过 Config.DataPath 指定 tessdata 父目录",
			lang, cfg.DataPath)
	}

	if cfg.PSM != 0 {
		C.ocr_baseapi_set_psm((*C.ocr_baseapi_t)(api), C.int(cfg.PSM))
	}

	return &Engine{api: api, level: level}, nil
}

// Close 释放引擎及其模型资源。关闭后引擎不可再使用，重复调用是安全的。
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil
	}
	if e.api != nil {
		C.ocr_baseapi_end((*C.ocr_baseapi_t)(e.api))
		C.ocr_baseapi_delete((*C.ocr_baseapi_t)(e.api))
		e.api = nil
	}
	e.closed = true
	return nil
}

// SetLevel 修改后续 Recognize 返回 Blocks 时使用的切分粒度。
func (e *Engine) SetLevel(level Level) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.level = level
}

// SetPageSegMode 设置页面分析模式（可在多次识别之间调整）。
func (e *Engine) SetPageSegMode(mode PageSegMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.closed {
		C.ocr_baseapi_set_psm((*C.ocr_baseapi_t)(e.api), C.int(mode))
	}
}

// Version 返回动态库内的 Tesseract 版本字符串；版本符号不可用时返回空串。
// 调用本方法同样会触发动态库加载。
func Version(libraryPath string) string {
	if err := ensureLoaded(libraryPath); err != nil {
		return ""
	}
	p := C.ocr_version()
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

// Recognize 对一张解码后的图片执行 OCR。
// 支持任意 image.Image；内部会按需转换为 8-bit RGBA（每像素 4 字节）。
func (e *Engine) Recognize(img image.Image) (*Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, errors.New("ocr: 引擎已关闭")
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, errors.New("ocr: 图片尺寸无效")
	}

	rgba := toRGBA(img)
	stride := rgba.Stride

	// Go 像素内存在此函数返回前始终有效，Tesseract 在 Recognize/Clear
	// 期间只做读取。
	C.ocr_baseapi_set_image(
		(*C.ocr_baseapi_t)(e.api),
		(*C.uchar)(unsafe.Pointer(&rgba.Pix[0])),
		C.int(w), C.int(h), C.int(stride),
	)
	// 识别完成后通知引擎释放对像素缓冲的内部引用。
	defer C.ocr_baseapi_clear((*C.ocr_baseapi_t)(e.api))

	if rc := C.ocr_baseapi_recognize((*C.ocr_baseapi_t)(e.api)); rc != 0 {
		return nil, fmt.Errorf("ocr: Recognize 失败，返回码 %d", int(rc))
	}

	// ---- 整页文本与平均置信度 ----
	res := &Result{}

	cText := C.ocr_baseapi_get_utf8((*C.ocr_baseapi_t)(e.api))
	if cText != nil {
		res.Text = C.GoString(cText)
		C.ocr_free_text(cText)
	}
	res.Text = trimSpace(res.Text)
	res.Confidence = float64(C.ocr_baseapi_mean_conf((*C.ocr_baseapi_t)(e.api)))

	// ---- 按粒度遍历文本块 ----
	level := e.level
	it := unsafe.Pointer(C.ocr_baseapi_get_iterator((*C.ocr_baseapi_t)(e.api)))
	if it == nil {
		// 某些页面（如纯空白）可能没有迭代器，返回空块列表。
		res.Blocks = []Block{}
		return res, nil
	}
	defer C.ocr_result_delete((*C.ocr_result_it_t)(it))

	cLevel := C.int(level)
	for {
		var left, top, right, bottom C.int
		hasBox := C.ocr_result_bbox((*C.ocr_result_it_t)(it), cLevel,
			&left, &top, &right, &bottom) != 0

		var text string
		cBlockText := C.ocr_result_text((*C.ocr_result_it_t)(it), cLevel)
		if cBlockText != nil {
			text = trimSpace(C.GoString(cBlockText))
			C.ocr_free_text(cBlockText)
		}

		// 置信度（百分比，0-100）
		conf := float64(C.ocr_result_conf((*C.ocr_result_it_t)(it), cLevel))

		block := Block{Text: text, Confidence: conf}
		if hasBox {
			block.BBox = Rect{
				X:      int(left),
				Y:      int(top),
				Width:  int(right) - int(left),
				Height: int(bottom) - int(top),
			}
		}
		res.Blocks = append(res.Blocks, block)

		if C.ocr_result_next((*C.ocr_result_it_t)(it), cLevel) == 0 {
			break
		}
	}

	return res, nil
}
