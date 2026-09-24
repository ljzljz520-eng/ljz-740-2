# goocr — Go OCR 引擎绑定

通过 cgo 桥接**本地 OCR 动态库**的 Go 绑定，提供：

- 初始化/销毁 OCR 模型引擎（模型目录、CPU/GPU、线程数）
- 识别图片（传入已编码的 PNG/JPEG/BMP… 字节或文件路径）
- 返回文本块（UTF-8 文本、置信度、四点四边形坐标）
- 处理不同操作系统的动态库命名与搜索差异（Linux/macOS/Windows）
- `ocr` 命令行工具：识别一张图片并输出 JSON

## 原生 ABI（`ocr.h`）

本地动态库只需实现 `ocr.h` 中声明的 6 个 C 函数：

| C 函数 | 作用 |
| --- | --- |
| `ocr_engine_create` | 加载模型、创建引擎，失败时返回错误串 |
| `ocr_engine_destroy` | 释放引擎 |
| `ocr_engine_recognize` | 识别图片字节，输出 `OCRBlock` 数组与数量 |
| `ocr_free_result` | 释放识别结果 |
| `ocr_version` | 后端版本字符串 |
| `ocr_free_string` | 释放库返回的错误字符串 |

每个 `OCRBlock` 包含：

```c
const char* text;        // UTF-8 文本
double      confidence;  // 0.0 ~ 1.0
double      xs[4], ys[4]; // 四边形 4 个角点（左上起顺时针，像素坐标）
```

可以用任意真实后端（PaddleOCR、Tesseract、RapidOCR、厂商 SDK…）封装实现该 ABI；
仓库自带一个无模型依赖的参考桩实现 `test/stub/ocr_stub.c`，用于编译与链路验证。

## 操作系统差异处理

### 库文件名

| 系统 | 首选名 | 备选 |
| --- | --- | --- |
| Linux | `libocr.so` | `libocr.so.1`、`libocr.so.0` |
| macOS | `libocr.dylib` | `libocr.1.dylib`、`libocr.0.dylib` |
| Windows | `ocr.dll` | `libocr.dll` |

实现位置：`libname_*.go`（编译期默认名）、`find_*.go`（运行时搜索路径）。

### 链接参数

`link_linux.go` / `link_darwin.go` / `link_windows.go` 分别给出平台 cgo 链接参数
（默认 `-locr`）。Windows 使用 MSVC 时需要导入库 `ocr.lib`，MinGW-w64 用
`libocr.dll.a`；若只有 DLL，可直接链接 DLL。

不使用默认链接名时，加 build tag 并自行传入：

```sh
go build -tags ocr_custom_lib .
CGO_LDFLAGS="-L/opt/ocr/lib -l:libocr.so.1" go build -tags ocr_custom_lib ./...
```

### 运行时库查找

`ocr.FindLibrary()` 按以下顺序查找动态库（跨平台、纯 Go）：

1. 环境变量 `OCR_LIBRARY_PATH`（冒号/分号分隔多个目录）
2. 可执行文件所在目录
3. 当前工作目录
4. 系统默认目录（Linux：`/usr/local/lib`、`/usr/lib`、multiarch 目录等；
   macOS：`/usr/local/lib`、`/opt/homebrew/lib` 等；Windows：`%WINDIR%\System32`）

找不到时返回包装了 `ocr.ErrLibraryNotFound` 的错误，并打印所有尝试过的路径。
运行时仍需保证动态库在加载器路径上（`LD_LIBRARY_PATH` /
`DYLD_LIBRARY_PATH` / `PATH`，或将库放到系统目录）。

## 快速开始

### 1. 准备动态库

真实后端：把实现 `ocr.h` 的库安装为系统名（如 `/usr/local/lib/libocr.so`）。

或用参考桩（无需模型即可体验完整链路）：

```sh
make stub        # 生成 test/stub/libocr.so（macOS 为 libocr.dylib）
```

### 2. 作为库使用

```go
package main

import (
    "fmt"
    "log"

    ocr "github.com/solo-manager/goocr"
)

func main() {
    eng, err := ocr.New(ocr.Config{
        ModelDir:   "./models",
        Device:     ocr.DeviceCPU, // 或 ocr.DeviceGPU
        GPUID:      0,
        NumThreads: 4,             // 0 表示由后端决定
    })
    if err != nil {
        log.Fatal(err)
    }
    defer eng.Close()

    res, err := eng.RecognizeFile("invoice.png")
    if err != nil {
        log.Fatal(err)
    }
    for _, b := range res.Blocks {
        fmt.Printf("[%.2f] %s\n", b.Confidence, b.Text)
    }
}
```

### 3. 命令行示例：识别图片并输出 JSON

```sh
# 用桩库一键演示
make example

# 真实使用
ocr -model ./models -image invoice.png -pretty
ocr -model ./models -image scan.jpg -device gpu -gpu-id 0 -threads 8
ocr -version          # 查看后端版本
ocr -check-library    # 只检查能否找到动态库
```

输出（`-pretty`）：

```json
{
  "backend": "stub-ocr/1.0.0",
  "image": "invoice.png",
  "count": 2,
  "result": {
    "blocks": [
      {
        "text": "Hello, OCR!",
        "confidence": 0.97,
        "points": [
          { "x": 12, "y": 8 },
          { "x": 202, "y": 8 },
          { "x": 202, "y": 40 },
          { "x": 12, "y": 40 }
        ]
      }
    ]
  }
}
```

CLI 参数：

| 参数 | 说明 | 默认 |
| --- | --- | --- |
| `-model` | 模型目录（必填） | — |
| `-image` | 图片路径（必填） | — |
| `-device` | `cpu` / `gpu` | `cpu` |
| `-gpu-id` | GPU 编号 | `0` |
| `-threads` | CPU 推理线程数，0 为后端默认 | `0` |
| `-pretty` | 缩进格式化 JSON | `false` |
| `-version` | 打印后端版本并退出 | — |
| `-check-library` | 定位动态库路径并退出 | — |

## 构建与测试

```sh
make build       # 构建桩库并链接编译全部包
make test-stub   # 基于桩库跑完整测试
make vet
make example
```

直接使用 go 工具链：

```sh
# 链接位于非标准目录的真实库
CGO_LDFLAGS="-L/opt/ocr/lib" \
LD_LIBRARY_PATH="/opt/ocr/lib" \
go test ./...

# 无 cgo 环境（类型与 FindLibrary 仍可用，FFI 调用返回明确错误）
CGO_ENABLED=0 go build ./...
```

测试默认在找不到动态库时**跳过**原生用例（纯 Go 的库查找用例始终运行）；
设置 `GOOCR_REQUIRE_LIB=1` 可将缺库视为失败。

## 线程安全与内存

- `Engine` 内部用互斥锁串行化所有 cgo 调用（底层推理会话通常不可并发）。
- 原生内存（引擎、block 数组、错误串）均在绑定层及时释放；Go 侧 `Block`
  为深拷贝，调用方无需手动管理。
- `Close()` 可重复调用；关闭后再识别会返回错误。

## 项目结构

```
ocr.h                 # 原生库需要实现的 C ABI
doc.go                # 包文档
types.go              # 纯 Go 公共类型（Engine/Config/Block/Result...）
bridge.go             # cgo 绑定：New/Close/Recognize/Version
link_{linux,darwin,windows}.go  # 各平台 cgo 链接参数
link_custom.go        # ocr_custom_lib tag：自带 CGO_LDFLAGS
libname{,_linux,_darwin,_windows,_other}.go  # 各平台默认库名
find.go + find_*.go   # 运行时跨平台动态库查找
stub_nocgo.go         # CGO_ENABLED=0 时的明确报错桩
cmd/ocr/              # 识别图片 -> JSON 的命令行
test/stub/ocr_stub.c  # 无模型依赖的参考动态库实现
```

## 故障排查

**启动时报 `error while loading shared libraries: libocr.so: cannot open shared object file`**

cgo 直接链接的库在进程启动、Go 代码执行之前由系统加载器装载。请用以下任一方式
让加载器找到它：

```sh
# Linux
export LD_LIBRARY_PATH=/opt/ocr/lib:$LD_LIBRARY_PATH
# 或安装到系统目录并刷新缓存
sudo cp /opt/ocr/lib/libocr.so /usr/local/lib/ && sudo ldconfig

# macOS
export DYLD_LIBRARY_PATH=/opt/ocr/lib:$DYLD_LIBRARY_PATH

# Windows (PowerShell)
$env:PATH = "C:\path\to\ocr;" + $env:PATH
```

也可以把库放在可执行文件同目录，或设置 `OCR_LIBRARY_PATH` 供
`ocr.FindLibrary()` / `ocr -check-library` 在 Go 层定位（注意：
`-check-library` 本身同样需要进程能先启动）。

**链接时报 `cannot find -locr`**

编译期找不到库：用 `CGO_LDFLAGS="-L<库目录>" go build ...` 指定搜索路径，
或使用 `-tags ocr_custom_lib` 自定义库名/路径。

**`CGO_ENABLED=0` 构建后调用报 `built with CGO_ENABLED=0`**

无 cgo 时只保留类型与 `FindLibrary`，实际识别必须重新用
`CGO_ENABLED=1` 且具备 C 编译器的环境构建。
