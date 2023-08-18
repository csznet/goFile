package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	bytes := make([]byte, stat.Size())

	// 读取文件内容
	_, err = io.ReadFull(file, bytes)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return nil, false
	}
	return bytes, true
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
