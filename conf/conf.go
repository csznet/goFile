package conf

type Info struct {
	Files []File
	Dirs  []Dir
}

type File struct {
	FileName string
	FilePath string
	IsZip    bool
	IsThumb  bool
}

type Dir struct {
	DirName string
	DirPath string
}

var (
	GoFilePort string
	GoFile     string
)
var GoCacheOption = false
