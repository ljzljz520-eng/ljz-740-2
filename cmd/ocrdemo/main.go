// 命令 ocrdemo 使用本地 Tesseract OCR 动态库识别一张图片，并把识别结果
// （整页文本、平均置信度、各文本块及其包围盒）输出为 JSON。
//
// 用法示例：
//
//	# 默认英文模型，识别 sample.png 并打印 JSON
//	ocrdemo -image sample.png
//
//	# 指定中文+英文、按文本行切分块、写入文件
//	ocrdemo -image photo.jpg -lang eng+chi_sim -level textline -out result.json
//
//	# 动态库不在系统搜索路径中时：
//	OCR_LIBRARY_PATH=/opt/homebrew/lib/libtesseract.dylib ocrdemo -image x.png
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"ocrbridge/ocr"
)

func main() {
	var (
		imagePath = flag.String("image", "", "待识别图片路径（必填；- 表示标准输入）")
		library   = flag.String("lib", os.Getenv("OCR_LIBRARY_PATH"),
			"OCR 动态库完整路径（默认读 OCR_LIBRARY_PATH，再尝试系统默认库名）")
		dataPath = flag.String("data", os.Getenv("TESSDATA_PREFIX"),
			"tessdata 目录路径，默认读 TESSDATA_PREFIX；留空则使用 Tesseract 内置路径")
		language  = flag.String("lang", "eng", "识别语言，如 eng、chi_sim、eng+chi_sim")
		levelName = flag.String("level", "block", "文本块粒度: block|paragraph|textline|word|symbol")
		psm       = flag.Int("psm", int(ocr.PSMAuto), "页面分析模式 PSM（0-13）")
		outPath   = flag.String("out", "", "JSON 输出文件路径（默认输出到标准输出）")
		showVer   = flag.Bool("version", false, "仅打印检测到的 Tesseract 版本后退出")
	)
	flag.Parse()

	level, ok := parseLevel(*levelName)
	if !ok {
		fatalf("未知的 -level %q（可选 block/paragraph/textline/word/symbol）", *levelName)
	}

	if *showVer {
		if v := ocr.Version(*library); v != "" {
			fmt.Println(v)
		} else {
			fatalf("无法获取 Tesseract 版本：请确认本地 OCR 动态库可被加载")
		}
		return
	}

	if *imagePath == "" {
		flag.Usage()
		os.Exit(2)
	}

	engine, err := ocr.New(ocr.Config{
		LibraryPath: *library,
		DataPath:    *dataPath,
		Language:    *language,
		Level:       level,
		PSM:         ocr.PageSegMode(*psm),
	})
	if err != nil {
		fatalf("%v", err)
	}
	defer engine.Close()

	result, err := engine.RecognizeFile(*imagePath)
	if err != nil {
		fatalf("识别失败: %v", err)
	}

	// 示例命令的输出结构：附带库版本与粒度信息，方便定位运行环境。
	output := struct {
		Image      string      `json:"image"`
		Language   string      `json:"language"`
		Level      string      `json:"level"`
		Library    string      `json:"library_version,omitempty"`
		Text       string      `json:"text"`
		Confidence float64     `json:"confidence"`
		Blocks     []ocr.Block `json:"blocks"`
	}{
		Image:      *imagePath,
		Language:   *language,
		Level:      level.String(),
		Library:    ocr.Version(""),
		Text:       result.Text,
		Confidence: result.Confidence,
		Blocks:     result.Blocks,
	}

	enc := json.NewEncoder(outputWriter(*outPath))
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fatalf("输出 JSON 失败: %v", err)
	}
}

func parseLevel(s string) (ocr.Level, bool) {
	switch s {
	case "block", "":
		return ocr.LevelBlock, true
	case "paragraph", "para":
		return ocr.LevelParagraph, true
	case "textline", "line":
		return ocr.LevelTextline, true
	case "word":
		return ocr.LevelWord, true
	case "symbol", "char":
		return ocr.LevelSymbol, true
	default:
		return 0, false
	}
}

func outputWriter(path string) *os.File {
	if path == "" {
		return os.Stdout
	}
	f, err := os.Create(path)
	if err != nil {
		fatalf("无法创建输出文件 %q: %v", path, err)
	}
	return f
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ocrdemo: "+format+"\n", args...)
	os.Exit(1)
}
