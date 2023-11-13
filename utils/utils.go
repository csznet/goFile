package utils

import (
	"archive/zip"
	"goFile/conf"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RemovePP(path string) string {
	return strings.ReplaceAll(path, "//", "/")
}

// Unzip 解压zip
func Unzip(src string) bool {
	OutPath := pathOutConv(src)
	src = conf.GoFile + src
	fr, err := zip.OpenReader(src)
	if err != nil {
		return false
	}
	defer fr.Close()
	//r.reader.file 是一个集合，里面包括了压缩包里面的所有文件
	for _, file := range fr.Reader.File {
		//判断文件该目录文件是否为文件夹
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(OutPath+file.Name, 0777)
			if err != nil {
				return false
			}
			continue
		}
		//为文件时，打开文件
		r, err := file.Open()
		if err == nil {
			//在对应的目录中创建相同的文件
			NewFile, _ := os.Create(OutPath + file.Name)
			//将内容复制
			io.Copy(NewFile, r)
			//关闭文件
			NewFile.Close()
		}
		r.Close()
	}
	return true
}

// 保存目录转换
func pathOutConv(path string) string {
	path = conf.GoFile + path
	fileSplit := strings.Split(path, "/")
	fileName := fileSplit[len(fileSplit)-1]
	OutPath := strings.TrimSuffix(path, fileName)
	return OutPath
}

// GetFile 远程下载
func GetFile(url, path string) bool {
	OutPath := pathOutConv(path)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	urlSplit := strings.Split(url, "/")
	fileName := urlSplit[len(urlSplit)-1]
	defer resp.Body.Close()
	// 创建一个文件用于保存
	out, err := os.Create(OutPath + fileName)
	if err != nil {
		return false
	}
	defer out.Close()
	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false
	}
	return true
}

// Exist 判断是否存在
func Exist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// In 判断是否在列表中
func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	//index的取值：[0,len(str_array)]
	if index < len(strArray) && strArray[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false
}

// GetFiles 获取文件列表
func GetFiles(path string) conf.Info {
	getFile, _ := filepath.Glob(path)
	var info conf.Info
	ZipList := []string{"zip", "gz"}
	for i := 0; i < len(getFile); i++ {
		im := getFile[i]
		if Exist(im) {
			s, _ := os.Stat(im)
			if s.IsDir() {
				var dir conf.Dir
				if conf.GoFile != "./" {
					dir.DirPath = strings.TrimPrefix(im, conf.GoFile)
				} else {
					dir.DirPath = im
				}
				dirSplit := strings.Split(im, "/")
				dir.DirName = dirSplit[len(dirSplit)-1]
				info.Dirs = append(info.Dirs, dir)
			} else {
				var file conf.File
				if conf.GoFile != "./" {
					file.FilePath = strings.TrimPrefix(im, conf.GoFile)
				} else {
					file.FilePath = im
				}
				filePath := strings.Split(im, "/")
				file.FileName = filePath[len(filePath)-1]
				file.IsZip = false
				strSplit := strings.Split(im, ".")
				if In(strSplit[len(strSplit)-1], ZipList) {
					file.IsZip = true
				}
				//file.FilePath = NewPath + im
				info.Files = append(info.Files, file)
			}
		}
	}
	return info
}
