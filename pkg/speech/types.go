package speech

import (
	"strings"
	"unicode"
)

// 评测模式类型
type ModeType int64

const (
	// 英文模式
	EngWord         ModeType = 0
	EngSentence     ModeType = 1
	EngParagraph    ModeType = 2
	EngFreeTalk     ModeType = 3
	EngWordCorrect  ModeType = 4
	EngScenario     ModeType = 5
	EngMultiBranch  ModeType = 6
	EngRealTimeWord ModeType = 7

	// 中文模式
	ChnWord         ModeType = 11
	ChnSentence     ModeType = 22
	ChnParagraph    ModeType = 33
	ChnFreeTalk     ModeType = 44
	ChnScenario     ModeType = 55
	ChnMultiBranch  ModeType = 66
	ChnRealTimeWord ModeType = 77
	ChnPinyin       ModeType = 88
)

// 评测模式配置（根据文档规则定义）
var modeRules = map[ModeType]struct {
	language    string // 语言类型（en/zh）
	minWords    int    // 最小单词/字符数（0表示不限）
	maxWords    int    // 最大单词/字符数（0表示不限）
	hasSpace    bool   // 是否包含空格（英文适用）
	hasBranch   bool   // 是否包含分支（如A./B.）
	hasPhonetic bool   // 是否包含音标（/.../）
	isPinyin    bool   // 是否为拼音
	isRealTime  bool   // 是否为实时多单词模式
}{
	// 英文模式
	EngWord: {
		language:    "en",
		minWords:    1,
		maxWords:    1,
		hasSpace:    false,
		hasPhonetic: true, // 支持音标
	},
	EngSentence: {
		language: "en",
		minWords: 2,
		maxWords: 30,
		hasSpace: true,
	},
	EngParagraph: {
		language: "en",
		minWords: 31,
		maxWords: 120,
		hasSpace: true,
	},
	EngFreeTalk: {
		language: "en",
		maxWords: 0, // 不限字数
	},
}

// DetectEvalMode 检测评测模式
func DetectEvalMode(input string) int64 {
	// 1. 判断语言类型
	language := detectLanguage(input)

	// 2. 预处理文本（去空格、标点，提取分支）
	cleanInput := strings.TrimSpace(input)
	hasBranch := strings.ContainsAny(cleanInput, "/\\|") // 分支分隔符
	// hasPhonetic := regexp.MustCompile(`/.*?/`).MatchString(cleanInput) // 检测音标格式
	// isPinyin := isPurePinyin(cleanInput)                               // 纯拼音检测

	// 3. 统计字数/单词数
	wordCount := countWords(cleanInput, language)

	// 4. 按语言过滤模式
	var possibleModes []ModeType
	for mode, rule := range modeRules {
		if rule.language != language {
			continue
		}
		possibleModes = append(possibleModes, mode)
	}

	// 5. 匹配具体模式
	for _, mode := range possibleModes {
		rule := modeRules[mode]

		// 检查字数限制
		if !checkWordCount(wordCount, rule.minWords, rule.maxWords) {
			continue
		}

		// 检查空格
		if rule.hasSpace && !strings.Contains(cleanInput, " ") {
			continue
		}
		if rule.hasSpace && rule.isRealTime { // 实时模式无空格
			continue
		}

		// 检查分支
		if rule.hasBranch && !hasBranch {
			continue
		}

		// 实时模式特殊处理（多个单词但无空格，按分支数判断）
		if rule.isRealTime {
			branchCount := len(strings.Split(cleanInput, "/")) // 假设分支用/分隔
			if branchCount >= 2 && wordCount == branchCount {  // 每个分支1个单词/汉字
				return int64(mode)
			}
			continue
		}

		// 自由说模式：字数不限且无分支/音标
		if mode == EngFreeTalk || mode == ChnFreeTalk {
			if !hasBranch {
				return int64(mode)
			}
			continue
		}

		// 其他模式默认匹配
		return int64(mode)
	}

	return 0 // 未匹配到模式
}

// 检测语言类型（英文/中文）
func detectLanguage(input string) string {
	for _, r := range input {
		if unicode.Is(unicode.Han, r) { // 检测汉字
			return "zh"
		}
		if unicode.IsLetter(r) && unicode.IsLower(r) { // 检测英文字母
			return "en"
		}
	}
	return "" // 无法识别的语言
}

// 统计单词数（英文按空格分割，中文按字符数）
func countWords(input, language string) int {
	if language == "en" {
		return len(strings.Fields(input)) // 按空格分割单词
	} else if language == "zh" {
		return len([]rune(input)) // 中文按字符数
	}
	return 0
}

// 检查字数是否符合范围（maxWords=0表示不限）
func checkWordCount(count, min, max int) bool {
	if min > 0 && count < min {
		return false
	}
	if max > 0 && count > max {
		return false
	}
	return true
}
