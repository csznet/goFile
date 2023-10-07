package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"goFile/conf"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
)

// ReadFile read file
func ReadFile(path string) ([]byte, bool) {
	// 读取配置文件
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open config file:", err)
		return nil, false
	}
	defer file.Close()

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file size:", err)
		return nil, false
	}

	// 分配足够的空间来存储文件内容
	freeBytes := make([]byte, stat.Size())

	// 读取文件内容
	_, err = io.ReadFull(file, freeBytes)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return nil, false
	}
	return freeBytes, true
}
func ExistFile(filePath string) bool {
	// 获取文件信息
	_, err := os.Stat(filePath)
	// 判断文件是否存在
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Println("Error:", err)
		return false
	} else {
		return true
	}
}

// NewImgThumb 生成Thumb
func NewImgThumb(fileContent []byte, thumbnailPath string) bool {
	buf := bytes.NewBuffer(fileContent)
	image, err := imaging.Decode(buf)
	if err != nil {
		fmt.Println(err)
		return false
	}
	thumbnail := imaging.Resize(image, 400, 0, imaging.Lanczos)
	// 获取文件夹路径
	thumbnailDir := filepath.Dir(thumbnailPath)

	// 检查文件夹是否存在，如果不存在则创建
	if _, err := os.Stat(thumbnailDir); os.IsNotExist(err) {
		err := os.MkdirAll(thumbnailDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return false
		}
	}
	err = imaging.Save(thumbnail, thumbnailPath)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
func RemovePP(path string) string {
	return strings.ReplaceAll(path, "//", "/")
}
func GetImgThumb(path, goCachePath string) bool {
	fileContent, getImgStatus := ReadFile(path)
	if getImgStatus {
		// //判断缓存文件夹是否存在（如果是默认文件夹则自动新建
		// if goCachePath == "/var/tmp/goFile/" {
		// 	_, err := os.Stat(goCachePath)
		// 	if os.IsNotExist(err) {
		// 		// 文件夹不存在，创建它
		// 		err := os.Mkdir(goCachePath, 0755) // 0755 是文件夹权限
		// 		if err != nil {
		// 			fmt.Println("Failed to create folder:", err)
		// 			return false
		// 		}
		// 	} else if err != nil {
		// 		// 其他错误
		// 		fmt.Println("Error:", err)
		// 		return false
		// 	}
		// }
		thumbnailPath := RemovePP(goCachePath + path)
		//判断thumb文件是否存在
		if ExistFile(thumbnailPath) {
			// 获取thumb文件信息
			fileInfo, err := os.Stat(thumbnailPath)
			if err != nil {
				fmt.Println("Error:", err)
				return false
			}
			// 获取文件最后修改时间
			thumbModTime := fileInfo.ModTime()
			fileInfo, err = os.Stat(path)
			if err != nil {
				fmt.Println("Error:", err)
				return false
			}
			//获取源文件最后修改时间
			modTime := fileInfo.ModTime()
			if thumbModTime.After(modTime) {
				//thumb是在源文件后建立的
				fmt.Println("thumbDate:", thumbModTime)
				fmt.Println(modTime)
				return true
			} else {
				return NewImgThumb(fileContent, thumbnailPath)
			}
		}
		return NewImgThumb(fileContent, thumbnailPath)
	} else {
		return false
	}
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

// Exists 判断是否存在
func exists(path string) bool {
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
	ImgList := []string{"jpg", "png"}
	for i := 0; i < len(getFile); i++ {
		im := getFile[i]
		if exists(im) {
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
				file.IsThumb = false
				if In(strSplit[len(strSplit)-1], ImgList) && conf.GoCacheOption {
					file.IsThumb = true
				}
				//file.FilePath = NewPath + im
				info.Files = append(info.Files, file)
			}
		}
	}
	return info
}
