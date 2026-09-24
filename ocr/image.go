package ocr

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif" // 注册 GIF 解码
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
)

// toRGBA 把任意 image.Image 转为紧凑的 *image.RGBA。
// 若源图本身已是 *image.RGBA 且 stride 恰好为 4*width 可直接复用；
// 否则逐像素转换。Tesseract 要求每像素连续 4 字节（RGBA）。
func toRGBA(src image.Image) *image.RGBA {
	if r, ok := src.(*image.RGBA); ok && r.Stride == 4*r.Rect.Dx() {
		return r
	}
	b := src.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, color.NRGBAModel.Convert(src.At(x, y)))
		}
	}
	return dst
}

// trimSpace 去除 Tesseract 返回文本两端多余空白（含换行）。
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// RecognizeReader 从 r 解码图片（PNG/JPEG/GIF）并执行识别。
func (e *Engine) RecognizeReader(r io.Reader) (*Result, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return e.Recognize(img)
}

// RecognizeBytes 从内存字节解码图片并执行识别。
func (e *Engine) RecognizeBytes(data []byte) (*Result, error) {
	return e.RecognizeReader(bytes.NewReader(data))
}

// RecognizeFile 读取并识别磁盘上的图片文件。
// 路径为 "-" 时从标准输入读取。
func (e *Engine) RecognizeFile(path string) (*Result, error) {
	if path == "-" {
		return e.RecognizeReader(os.Stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return e.RecognizeReader(f)
}
