package main

import (
	"flag"
	"fmt"
	"goFile/assets"
	"goFile/conf"
	"goFile/i18n"
	"goFile/utils"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	reader = false
	cLang  i18n.LangType
)

// LangMiddleware i18n
func LangMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the language from the request header
		lang := c.GetHeader("Accept-Language")
		var langType i18n.LangType
		// Set the default language to English
		if strings.Contains(lang, "zh-CN") {
			langType = i18n.ZH
		} else {
			langType = i18n.EN
		}
		if cLang != langType {
			cLang = langType
		}
		// Call the next handler
		c.Next()
	}
}

func translate(key string) string {
	return i18n.Translate(key, cLang)
}

// Web Serve
func web() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(LangMiddleware())
	r.SetFuncMap(template.FuncMap{
		"t": translate,
	})
	r.SetHTMLTemplate(template.Must(template.New("").Funcs(r.FuncMap).ParseFS(assets.Templates, "templates/*")))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"info":   utils.GetFiles(conf.GoFile + "*"),
			"path":   "",
			"Lang":   c.GetString("Lang"),
			"reader": reader,
		})
	})
	r.GET("/view/*path", func(c *gin.Context) {
		cPath := strings.Replace(c.Param("path"), "/", "", 1)
		fileName := filepath.Base(cPath)
		c.FileAttachment(conf.GoFile+cPath, fileName)
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
			if conf.GoFile != "./" {
				goPath = conf.GoFile + "/" + cPath + "/*"
			}
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"info":   utils.GetFiles(goPath),
				"path":   cPath + "/",
				"prev":   prev,
				"reader": reader,
			})
		}
	})
	//非阅读模式
	if !reader {
		// r.POST("/get", func(c *gin.Context) {
		// 	go utils.GetFile(c.PostForm("url"), c.PostForm("path"))
		// 	cPath := strings.Replace(c.Param("path"), "/", "", 1)
		// 	if cPath == "/" {
		// 		cPath = ""
		// 	}
		// 	url := cPath
		// 	if len(cPath) == 0 {
		// 		url = "/"
		// 	} else {
		// 		url = "/d/" + url
		// 	}
		// 	c.HTML(http.StatusOK, "msg.tmpl", gin.H{
		// 		"msg":   translate("scrDown"),
		// 		"title": translate("rt"),
		// 		"url":   url,
		// 	})

		// })
		r.POST("/do/upload/*path", func(c *gin.Context) {
			cPath := strings.Replace(c.Param("path"), "/", "", 1)
			if cPath == "/" {
				cPath = ""
			}
			file, err := c.FormFile("file")
			Stat := translate("sc")
			if err != nil {
				Stat = translate("fl")
			}
			c.SaveUploadedFile(file, conf.GoFile+cPath+file.Filename)
			url := cPath
			if len(cPath) == 0 {
				url = "/"
			} else {
				url = "/d/" + url
			}
			c.HTML(http.StatusOK, "msg.tmpl", gin.H{
				"msg":   translate("upFile") + Stat,
				"title": translate("rt"),
				"url":   url,
			})
		})
		// 新建文件
		r.POST("/do/newfile", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			ok := true
			file := filepath.Join(conf.GoFile+c.PostForm("path"), c.PostForm("filename"))
			//判断文件是否存在
			if !utils.Exist(file) {
				ok = false
			}
			f, err := os.Create(file)
			defer f.Close()
			if err != nil {
				ok = false
			}
			c.JSON(http.StatusOK, gin.H{
				"stat": ok,
			})
		})
		// 新建文件夹
		r.POST("/do/newdir", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			ok := true
			dir := filepath.Join(conf.GoFile+c.PostForm("path"), c.PostForm("dirname"))
			if err := os.Mkdir(dir, 0755); err != nil {
				ok = false
			}
			c.JSON(http.StatusOK, gin.H{
				"stat": ok,
			})
		})
		//解压文件
		r.POST("/do/unzip", func(c *gin.Context) {
			path := c.PostForm("path")
			pathSplit := strings.Split(path, ".")
			fileType := pathSplit[len(pathSplit)-1]
			ok := false
			switch fileType {
			case "zip":
				ok = utils.Unzip(path)
			case "gz":
				cmd := exec.Command("tar", "-zxvf", path)
				err := cmd.Run()
				if err == nil {
					ok = true
				}
			}
			c.JSON(http.StatusOK, gin.H{"stat": ok})
		})

		//保存代码
		r.POST("/do/save/", func(c *gin.Context) {
			file, err := os.OpenFile(c.PostForm("path"), os.O_WRONLY|os.O_TRUNC, 0644)
			ok := true
			if err != nil {
				ok = false
			}
			data := c.PostForm("data")
			if len(data) > 0 && data[len(data)-1] == '\n' {
				data = data[:len(data)-1] //去掉新行
			}
			_, err = file.WriteString(data)
			defer file.Close()
			if err != nil && ok {
				ok = false
			}
			c.JSON(http.StatusOK, gin.H{
				"stat": ok,
			})
		})
		//编辑代码
		r.POST("/edite/", func(c *gin.Context) {
			file, _ := os.Open(c.PostForm("path"))
			data, _ := io.ReadAll(file)
			defer file.Close()
			c.HTML(http.StatusOK, "editor.tmpl", gin.H{
				"data": string(data),
				"path": c.PostForm("path"),
			})
		})
		//删除文件/文件夹
		r.POST("/do/rm", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			path := conf.GoFile + c.PostForm("path")
			err := os.RemoveAll(path)
			Stat := true
			if err != nil {
				Stat = true
			}
			c.JSON(http.StatusOK, gin.H{
				"stat": Stat,
			})
		})
	}
	//监听端口默认为8080
	r.Run("0.0.0.0:" + conf.GoFilePort)
}

func init() {
	flag.StringVar(&conf.GoFile, "path", "./", "goFile path")
	flag.StringVar(&conf.GoFilePort, "port", "8089", "goFile web port")
	readerPtr := flag.Bool("r", false, "Enable reader")
	flag.Parse()
	if *readerPtr {
		reader = true
	}
}
func main() {
	// 获取当前工作目录
	if conf.GoFile == "./" {
		cwd, err := os.Getwd()
		if err == nil {
			conf.GoFile = cwd
		}
	} else {
		conf.GoFile = strings.Replace(conf.GoFile, "./", "", 1)
	}
	if conf.GoFile[len(conf.GoFile)-1] != '/' {
		conf.GoFile = conf.GoFile + "/"
	}
	fmt.Println("Run Directory:" + conf.GoFile)
	fmt.Println("goFile Port is " + conf.GoFilePort)
	web()
}
