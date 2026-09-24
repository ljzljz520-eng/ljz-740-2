// 包 ocr 是本地 Tesseract OCR 动态库的 Go 绑定。
//
// 该绑定不在编译期链接 Tesseract，而是在运行时通过 dlopen（Linux/macOS）
// 或 LoadLibrary（Windows）加载动态库并解析符号，因此：
//
//   - 构建时无需安装 Tesseract 开发头文件，只需 C 编译器（cgo）；
//   - 运行时需要目标操作系统上存在 Tesseract 动态库及训练数据（tessdata）；
//   - 不同系统的库名差异（libtesseract.so / libtesseract.dylib / tesseract.dll）
//     已在内部处理，也可通过 Config.LibraryPath 或环境变量 OCR_LIBRARY_PATH
//     显式指定。
package ocr

// Level 表示 OCR 结果的切分粒度，取值与 Tesseract 的 PageIteratorLevel 对应。
type Level int

const (
	// LevelBlock 文本块（默认，通常对应段落/文本区域）。
	LevelBlock Level = iota
	// LevelParagraph 段落。
	LevelParagraph
	// LevelTextline 文本行。
	LevelTextline
	// LevelWord 单词。
	LevelWord
	// LevelSymbol 单个字符符号。
	LevelSymbol
)

// String 返回粒度的可读名称。
func (l Level) String() string {
	switch l {
	case LevelBlock:
		return "block"
	case LevelParagraph:
		return "paragraph"
	case LevelTextline:
		return "textline"
	case LevelWord:
		return "word"
	case LevelSymbol:
		return "symbol"
	default:
		return "unknown"
	}
}

// PageSegMode 是 Tesseract 的页面分析模式（PSM），取值与官方定义一致。
type PageSegMode int

const (
	PSMOSDOnly             PageSegMode = 0  // 仅方向和脚本检测
	PSMAutoOSD             PageSegMode = 1  // 自动分页 + OSD
	PSMAutoOnly            PageSegMode = 2  // 自动分页，无 OSD/OSD
	PSMAuto                PageSegMode = 3  // 全自动分页（默认）
	PSMSingleColumn        PageSegMode = 4  // 单列文本
	PSMSingleBlockVertText PageSegMode = 5  // 单一垂直对齐文本块
	PSMSingleBlock         PageSegMode = 6  // 单一统一文本块
	PSMSingleLine          PageSegMode = 7  // 单行文本
	PSMSingleWord          PageSegMode = 8  // 单个单词
	PSMCircleWord          PageSegMode = 9  // 圆形排列的单个单词
	PSMSingleChar          PageSegMode = 10 // 单个字符
	PSMSparseText          PageSegMode = 11 // 稀疏文本（无特定顺序）
	PSMSparseTextOSD       PageSegMode = 12 // 稀疏文本 + OSD
	PSMRawLine             PageSegMode = 13 // 原始行，绕过分字器
)

// Rect 是文本块在原图中的轴对齐包围盒，像素坐标原点位于左上角。
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Block 表示一个识别结果文本块：文本内容、置信度（0-100）和包围盒。
type Block struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	BBox       Rect    `json:"bbox"`
}

// Result 是一次图片识别的完整结果。
type Result struct {
	// Text 为整页拼接后的纯文本（已去除首尾空白）。
	Text string `json:"text"`
	// Confidence 为整页平均置信度，范围 0-100，越高越好。
	Confidence float64 `json:"confidence"`
	// Blocks 为按 Config.Level 切分得到的文本块列表。
	Blocks []Block `json:"blocks"`
}

// Config 是 OCR 引擎的初始化配置。
type Config struct {
	// LibraryPath 为 OCR 动态库的完整路径。
	// 留空时依次查找环境变量 OCR_LIBRARY_PATH 与各系统默认候选名。
	LibraryPath string

	// DataPath 为 tessdata 目录的路径（直接包含 eng.traineddata 的那一层），
	// 例如 "/usr/share/tesseract-ocr/5/tessdata"。
	// 留空（推荐）时由 Tesseract 按内置默认路径以及 TESSDATA_PREFIX 环境
	// 变量自行查找。
	DataPath string

	// Language 为识别语言，如 "eng"、"chi_sim"、"eng+chi_sim"。
	// 留空默认为 "eng"。
	Language string

	// Level 指定返回 Blocks 的切分粒度，留空（0 值）时按 LevelBlock 处理。
	Level Level

	// PSM 指定页面分析模式，留空（0 值）时使用 Tesseract 默认（PSMAuto）。
	PSM PageSegMode
}
