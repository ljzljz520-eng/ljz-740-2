package ocr

import (
	"image"
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	cases := map[Level]string{
		LevelBlock:     "block",
		LevelParagraph: "paragraph",
		LevelTextline:  "textline",
		LevelWord:      "word",
		LevelSymbol:    "symbol",
	}
	for l, want := range cases {
		if got := l.String(); got != want {
			t.Fatalf("Level(%d).String() = %q, want %q", int(l), got, want)
		}
	}
}

func TestRecognizeInvalidImage(t *testing.T) {
	eng, err := New(Config{})
	if err != nil {
		t.Skipf("当前环境无法加载本地 OCR 库，跳过: %v", err)
	}
	defer eng.Close()

	if _, err := eng.Recognize(image.NewRGBA(image.Rect(0, 0, 0, 10))); err == nil {
		t.Fatal("期望对空图片返回错误")
	}
}

func TestRecognizeBlankImage(t *testing.T) {
	eng, err := New(Config{})
	if err != nil {
		t.Skipf("当前环境无法加载本地 OCR 库，跳过: %v", err)
	}
	defer eng.Close()

	// 全白图片：识别结果文本应为空，接口本身不应报错。
	res, err := eng.Recognize(image.NewRGBA(image.Rect(0, 0, 64, 32)))
	if err != nil {
		t.Fatalf("识别空白图片失败: %v", err)
	}
	if strings.TrimSpace(res.Text) != "" {
		t.Fatalf("空白图片不应识别出文本，得到 %q", res.Text)
	}
	if res.Blocks == nil {
		t.Fatal("Blocks 应为非 nil 切片")
	}
}

func TestCloseIdempotent(t *testing.T) {
	eng, err := New(Config{})
	if err != nil {
		t.Skipf("当前环境无法加载本地 OCR 库，跳过: %v", err)
	}
	if err := eng.Close(); err != nil {
		t.Fatal(err)
	}
	if err := eng.Close(); err != nil {
		t.Fatalf("重复 Close 应返回 nil，得到 %v", err)
	}
	if _, err := eng.Recognize(image.NewRGBA(image.Rect(0, 0, 4, 4))); err == nil {
		t.Fatal("Close 后再 Recognize 应返回错误")
	}
}
