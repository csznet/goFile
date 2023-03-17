package main

import (
	"archive/zip"
	"flag"
	"github.com/gin-gonic/gin"
	"goFile/conf"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var goFile, goFilePort string

func web() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"info": getFiles(goFile + "*"),
			"path": "",
		})
	})
	r.GET("/view/*path", func(c *gin.Context) {
		cPath := strings.Replace(c.Param("path"), "/", "", 1)
		c.File(goFile + cPath)
	})
	r.POST("/get", func(c *gin.Context) {
		go getFile(c.PostForm("url"), c.PostForm("path"))
		cPath := strings.Replace(c.Param("path"), "/", "", 1)
		if cPath == "/" {
			cPath = ""
		}
		url := cPath
		if len(cPath) == 0 {
			url = "/"
		} else {
			url = "/d/" + url
		}
		c.HTML(http.StatusOK, "msg.tmpl", gin.H{
			"msg":   "远程下载已加入队列",
			"title": "返回",
			"url":   url,
		})
	})
	r.POST("/do/upload/*path", func(c *gin.Context) {
		cPath := strings.Replace(c.Param("path"), "/", "", 1)
		if cPath == "/" {
			cPath = ""
		}
		file, err := c.FormFile("file")
		Stat := "成功"
		if err != nil {
			Stat = "失败"
		}
		c.SaveUploadedFile(file, goFile+cPath+file.Filename)
		url := cPath
		if len(cPath) == 0 {
			url = "/"
		} else {
			url = "/d/" + url
		}
		c.HTML(http.StatusOK, "msg.tmpl", gin.H{
			"msg":   "上传文件" + Stat,
			"title": "返回",
			"url":   url,
		})
	})
	r.POST("/do/unzip", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{
			"stat": Unzip(c.PostForm("path")),
		})
	})
	r.POST("/do/rm", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		path := goFile + c.PostForm("path")
		err := os.RemoveAll(path)
		Stat := true
		if err != nil {
			Stat = true
		}
		c.JSON(http.StatusOK, gin.H{
			"stat": Stat,
		})
	})
	r.GET("/d/*path", func(c *gin.Context) {
		//防止提权
		if c.Param("path") == "/" {
			c.Redirect(http.StatusMovedPermanently, "/")
		} else {
			cPath := strings.Replace(c.Param("path"), "/", "", 1)
			cPath = strings.TrimSuffix(cPath, "/")
			pathSplice := strings.Split(cPath, "/")
			prev := strings.TrimSuffix(cPath, "/"+pathSplice[len(pathSplice)-1])
			if len(pathSplice) == 1 {
				prev = "/"
			} else {
				prev = "/d/" + prev
			}
			goPath := cPath + "/*"
			if goFile != "./" {
				goPath = goFile + "/" + cPath + "/*"
			}
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"info": getFiles(goPath),
				"path": cPath + "/",
				"prev": prev,
			})
		}
	})
	//监听端口默认为8080
	r.Run("0.0.0.0:" + goFilePort)
}
func getFile(url, path string) bool {
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
func pathOutConv(path string) string {
	path = goFile + path
	fileSplit := strings.Split(path, "/")
	fileName := fileSplit[len(fileSplit)-1]
	OutPath := strings.TrimSuffix(path, fileName)
	return OutPath
}
func Unzip(src string) bool {
	OutPath := pathOutConv(src)
	src = goFile + src
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

// Exists 判断是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
func in(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	//index的取值：[0,len(str_array)]
	if index < len(strArray) && strArray[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false
}
func getFiles(path string) conf.Info {
	getFile, _ := filepath.Glob(path)
	var info conf.Info
	ZipList := []string{"zip"}
	for i := 0; i < len(getFile); i++ {
		im := getFile[i]
		if Exists(im) {
			s, _ := os.Stat(im)
			if s.IsDir() {
				var dir conf.Dir
				if goFile != "./" {
					dir.DirPath = strings.TrimPrefix(im, goFile)
				} else {
					dir.DirPath = im
				}
				dirSplit := strings.Split(im, "/")
				dir.DirName = dirSplit[len(dirSplit)-1]
				info.Dirs = append(info.Dirs, dir)
			} else {
				var file conf.File
				if goFile != "./" {
					file.FilePath = strings.TrimPrefix(im, goFile)
				} else {
					file.FilePath = im
				}
				filePath := strings.Split(im, "/")
				file.FileName = filePath[len(filePath)-1]
				file.IsZip = false
				strSplit := strings.Split(im, ".")
				if in(strSplit[len(strSplit)-1], ZipList) {
					file.IsZip = true
				}
				//file.FilePath = NewPath + im
				info.Files = append(info.Files, file)
			}
		}
	}
	return info
}
func init() {
	flag.StringVar(&goFile, "path", "./", "goFile path")
	flag.StringVar(&goFilePort, "port", "8089", "goFile web port")
	if goFile != "./" {
		goFile = strings.Replace(goFile, "./", "", 1)
	}
	if goFile[len(goFile)-1] != '/' {
		goFile = goFile + "/"
	}
}
func main() {
	flag.Parse()
	web()
}
