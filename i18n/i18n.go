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
	"scrDown": {
		EN: "Remote download has been queued",
		ZH: "远程下载已加入队列",
	},
	"del": {
		EN: "delete",
		ZH: "删除",
	},
	"edit": {
		EN: "edit",
		ZH: "编辑",
	},
	"unzip": {
		EN: "decompression",
		ZH: "解压",
	},
	"submit": {
		EN: "submit",
		ZH: "提交",
	},
	"cUpFile": {
		EN: "Upload files to this directory",
		ZH: "上传文件到此目录",
	},
	"nFile": {
		EN: "New File",
		ZH: "新建文件",
	},
	"nDir": {
		EN: "New Dir",
		ZH: "新建文件夹",
	},
	"reqFail": {
		EN: "request failure",
		ZH: "请求失败",
	},
	"view": {
		EN: "View",
		ZH: "预览",
	},
}

func Translate(key string, lang LangType) string {
	if translations[key] != nil {
		return translations[key][lang]
	}
	return key
}
