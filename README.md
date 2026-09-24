# ocrbridge — Go OCR 引擎绑定（本地 Tesseract 动态库桥接）

通过 cgo 在 **运行时** 桥接本地 [Tesseract](https://github.com/tesseract-ocr/tesseract)
OCR 动态库，提供：

- 初始化模型（指定 tessdata 目录与语言，如 `eng` / `chi_sim` / `eng+chi_sim`）；
- 识别 `image.Image` / 文件 / `io.Reader`（PNG、JPEG、GIF）；
- 返回整页文本、平均置信度，以及按块 / 段落 / 行 / 词 / 字符切分的
  **文本块列表**，每块含文本、置信度（0–100）与像素级包围盒；
- JSON 输出的示例命令 `ocrdemo`。

## 与常见 cgo 绑定的区别

本包 **不在编译期链接 Tesseract**：

- 构建时 **不需要** 安装 Tesseract 开发头文件 / 导入库，只需一个 C 编译器；
- 运行时用 `dlopen`（Linux/macOS）或 `LoadLibrary`（Windows）加载动态库并
  解析符号，库缺失时报清晰的错误而不是让整个程序无法启动；
- 不同操作系统的库名差异（`libtesseract.so` / `libtesseract.dylib` /
  `tesseract.dll`）已在各平台的候选列表中处理，也支持显式指定。

代价：必须设置 `CGO_ENABLED=1`（cgo 无法关闭）。

## 目录结构

| 文件 | 作用 |
| --- | --- |
| `ocr/bridge.h` | C 桥接层：Tesseract C ABI 函数指针声明 + 薄包装（不依赖官方头文件） |
| `ocr/loader_posix.h` | Linux/macOS 共用的 `dlopen` 加载逻辑 |
| `ocr/lib_linux.c` | Linux 候选库名：`libtesseract.so[.3/.4/.5]` 等 |
| `ocr/lib_darwin.c` | macOS 候选库名：`libtesseract[.4/.5].dylib` 及 Homebrew/MacPorts 路径 |
| `ocr/lib_windows.c` | Windows `LoadLibrary`：`tesseract.dll`、`tesseract54.dll` 及常见安装路径 |
| `ocr/engine.go` | 引擎：`New` / `Close` / `Recognize` / `SetPageSegMode` / `Version` |
| `ocr/types.go` | `Config` / `Result` / `Block` / `Rect` / `Level` / `PageSegMode` |
| `ocr/image.go` | 图片解码与 RGBA 转换、文件 / Reader / 字节入口 |
| `cmd/ocrdemo` | 示例命令：识别图片并输出 JSON |

## 前置条件

1. 构建机有 C 编译器（Linux 上为 gcc，macOS 为 Xcode Command Line Tools，
   Windows 上为 MinGW-w64）。
2. 运行机已安装 Tesseract 动态库与至少一种语言的训练数据：

   - Debian/Ubuntu：`apt install libtesseract5 tesseract-ocr tesseract-ocr-eng`
     （中文再加 `tesseract-ocr-chi-sim`）
   - macOS：`brew install tesseract tesseract-lang`
   - Windows：安装 [UB-Mannheim 构建](https://github.com/UB-Mannheim/tesseract/wiki)，
     并确保 `tesseract.dll` 在程序目录或 `PATH` 中。

## 库的查找顺序

1. `Config.LibraryPath`（或 `-lib` 命令行参数）；
2. 环境变量 `OCR_LIBRARY_PATH`；
3. 当前操作系统的默认候选名 / 常见安装路径。

训练数据目录由 Tesseract 自行解析（内置路径或 `TESSDATA_PREFIX`）；
也可通过 `Config.DataPath` 显式指定 **tessdata 目录本身**
（即直接包含 `eng.traineddata` 的那一层）。

## 作为库使用

```go
import "ocrbridge/ocr"

eng, err := ocr.New(ocr.Config{
    Language: "eng+chi_sim",   // 留空默认 eng
    Level:    ocr.LevelWord,   // 按单词返回文本块
})
if err != nil {
    log.Fatal(err)
}
defer eng.Close()

res, err := eng.RecognizeFile("photo.png")
if err != nil {
    log.Fatal(err)
}

fmt.Println(res.Text)            // 整页文本
fmt.Println(res.Confidence)      // 整页平均置信度 0-100
for _, b := range res.Blocks {
    fmt.Printf("%-20s conf=%6.2f box=%+v\n", b.Text, b.Confidence, b.BBox)
}
```

也可以直接识别解码后的图片：

```go
f, _ := os.Open("scan.jpg")
defer f.Close()
res, err := eng.RecognizeReader(f)
```

## 示例命令 ocrdemo

构建：

```bash
go build -o ocrdemo ./cmd/ocrdemo
```

识别一张图片并输出 JSON（默认英文、按文本块）：

```bash
$ ./ocrdemo -image sample.png
{
  "image": "sample.png",
  "language": "eng",
  "level": "block",
  "library_version": "5.3.0",
  "text": "HI OF 42",
  "confidence": 86,
  "blocks": [
    {
      "text": "HI OF 42",
      "confidence": 86.84,
      "bbox": { "x": 20, "y": 20, "width": 324, "height": 42 }
    }
  ]
}
```

常用参数：

```text
-image   图片路径（- 表示标准输入）
-lang    语言，如 chi_sim / eng+chi_sim（默认 eng）
-level   block | paragraph | textline | word | symbol（默认 block）
-psm     页面分析模式 0-13（默认 3，全自动）
-data    tessdata 目录（默认取 TESSDATA_PREFIX）
-lib     动态库完整路径（默认取 OCR_LIBRARY_PATH）
-out     JSON 输出文件（默认标准输出）
-version 仅打印 Tesseract 版本
```

按文本行切分、中英双语、写入文件：

```bash
./ocrdemo -image photo.jpg -lang eng+chi_sim -level textline -out result.json
```

## 线程模型

单个 `*Engine` 内部带互斥锁，可被多个 goroutine 共享，但识别会串行执行
（Tesseract 单实例不可并发）。需要真正并行时请创建多个 `Engine`。

## 跨平台编译

```bash
# Windows（需要 mingw-w64）
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    go build -o ocrdemo.exe ./cmd/ocrdemo
```
