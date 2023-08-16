package utils

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"os"
)

// ReadFile read file
func ReadFile(path string) (*os.File, bool) {
	// 读取配置文件
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open config file:", err)
		return file, false
	}
	defer file.Close()

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Failed to get file size:", err)
		return file, false
	}

	// 分配足够的空间来存储文件内容
	bytes := make([]byte, stat.Size())

	// 读取文件内容
	_, err = io.ReadFull(file, bytes)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return file, false
	}
	return file, true
}

// GetImgThumb img thumb
func GetImgThumb(path, goCachePath string) bool {
	getImg, getImgStatus := ReadFile(path)
	if getImgStatus {
		buf := &bytes.Buffer{}
		_, err := io.Copy(buf, getImg)
		if err != nil {
			fmt.Println("can't copy file:", err)
			return false
		}
		image, err := imaging.Decode(buf)
		if err != nil {
			fmt.Println(err)
			return false
		}
		//生成缩略图，尺寸150*200，并保持到为文件2.jpg
		image = imaging.Resize(image, 150, 200, imaging.Lanczos)
		err = imaging.Save(image, "d:/2.jpg")
		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	} else {
		return false
	}
}
