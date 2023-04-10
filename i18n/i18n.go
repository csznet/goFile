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
	"sc": {
		EN: "Success",
		ZH: "成功",
	},
	"fl": {
		EN: "Fail",
		ZH: "失败",
	},
	"rt": {
		EN: "Return",
		ZH: "返回",
	},
	"upFile": {
		EN: "Upload files",
		ZH: "上传文件",
	},
}

func Translate(key string, lang LangType) string {
	if translations[key] != nil {
		return translations[key][lang]
	}
	return key
}
