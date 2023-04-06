package i18n

type LangType int

const (
	EN LangType = iota
	ZH
)

var translations = map[string]map[LangType]string{
	"cd": {
		EN: "Current directory",
		ZH: "当前目录",
	},
	"rl": {
		EN: "Return to the previous level",
		ZH: "返回上一级",
	},
	"dls": {
		EN: "Directory List",
		ZH: "目录列表",
	},
	"fls": {
		EN: "File List",
		ZH: "文件列表",
	},
}

func Translate(key string, lang LangType) string {
	if translations[key] != nil {
		return translations[key][lang]
	}
	return key
}
