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
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	reader       = false
	uploader     = false
	templateSets map[i18n.LangType]*template.Template
)

// initTemplates pre-compiles one template set per language so each request
// gets a thread-safe, language-bound "t" function without using a shared global.
func initTemplates() {
	templateSets = make(map[i18n.LangType]*template.Template)
	for _, lang := range []i18n.LangType{i18n.EN, i18n.ZH} {
		l := lang
		templateSets[l] = template.Must(
			template.New("").Funcs(template.FuncMap{
				"t": func(key string) string {
					return i18n.Translate(key, l)
				},
			}).ParseFS(assets.Templates, "templates/*"),
		)
	}
}

// getLang returns the language stored in the request context by LangMiddleware.
func getLang(c *gin.Context) i18n.LangType {
	if lang, ok := c.Get("lang"); ok {
		return lang.(i18n.LangType)
	}
	return i18n.ZH
}

// LangMiddleware detects the request language and stores it in the context.
func LangMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		langType := i18n.ZH
		if strings.Contains(lang, "en") && !strings.Contains(lang, "zh") {
			langType = i18n.EN
		}
		c.Set("lang", langType)
		c.Next()
	}
}

// formatBytes returns a human-readable byte size string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// renderHTML executes the named template with the per-request language set.
func renderHTML(c *gin.Context, name string, data gin.H) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	lang := getLang(c)
	data["htmlLang"]    = map[i18n.LangType]string{i18n.ZH: "zh-CN", i18n.EN: "en"}[lang]
	data["allowUpload"] = !reader || uploader
	if total, free := utils.DiskUsage(conf.GoFile); total > 0 {
		used := total - free
		data["diskPct"]   = int(used * 100 / total)
		data["diskFree"]  = formatBytes(free)
		data["diskTotal"] = formatBytes(total)
	}
	if err := templateSets[lang].ExecuteTemplate(c.Writer, name, data); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// logAction prints a single-line access log entry.
func logAction(c *gin.Context, typ, path string) {
	fmt.Printf("%s  [%s]  %s  %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		typ,
		c.ClientIP(),
		path,
	)
}

// urlForPath returns the redirect URL for a given relative directory path.
func urlForPath(cPath string) string {
	if len(cPath) == 0 {
		return "/"
	}
	return "/d/" + cPath
}

// chunkDir returns the temp directory for storing upload chunks.
// Returns empty string if fileId is invalid (path traversal attempt).
func chunkDir(fileId string) string {
	base := filepath.Join(os.TempDir(), "goFile-chunks")
	dir := filepath.Clean(filepath.Join(base, fileId))
	if !strings.HasPrefix(dir, base+string(filepath.Separator)) {
		return ""
	}
	return dir
}

// Web Serve
func web() {
	initTemplates()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LangMiddleware())

	r.GET("/", func(c *gin.Context) {
		logAction(c, "浏览", "/")
		renderHTML(c, "index.tmpl", gin.H{
			"info":   utils.GetFiles(conf.GoFile + "*"),
			"path":   "",
			"reader": reader,
		})
	})

	r.GET("/view/*path", func(c *gin.Context) {
		absPath := filepath.Join(conf.GoFile, filepath.Clean(c.Param("path")))
		if !utils.IsPathSafe(absPath) {
			c.Status(http.StatusForbidden)
			return
		}
		logAction(c, "查看", c.Param("path"))
		c.File(absPath)
	})

	r.GET("/download/*path", func(c *gin.Context) {
		cPath := filepath.Clean(c.Param("path"))
		absPath := filepath.Join(conf.GoFile, cPath)
		if !utils.IsPathSafe(absPath) {
			c.Status(http.StatusForbidden)
			return
		}
		logAction(c, "下载", cPath)
		c.FileAttachment(absPath, filepath.Base(cPath))
	})

	r.GET("/d/*path", func(c *gin.Context) {
		rawPath := c.Param("path")
		if rawPath == "/" {
			c.Redirect(http.StatusMovedPermanently, "/")
			return
		}
		logAction(c, "浏览", rawPath)
		cPath := strings.TrimPrefix(rawPath, "/")
		cPath = strings.TrimSuffix(cPath, "/")
		prev := utils.GetPrevPath(cPath)
		goPath := filepath.Join(conf.GoFile, cPath) + "/*"
		renderHTML(c, "index.tmpl", gin.H{
			"info":   utils.GetFiles(goPath),
			"path":   cPath + "/",
			"prev":   prev,
			"reader": reader,
		})
	})

	// 上传路由（普通模式 + -ru 模式均开放）
	if !reader || uploader {
		r.POST("/do/upload/*path", func(c *gin.Context) {
			cPath := strings.Trim(c.Param("path"), "/")
			lang := getLang(c)

			file, err := c.FormFile("file")
			if err != nil {
				renderHTML(c, "msg.tmpl", gin.H{
					"msg":   i18n.Translate("upFile", lang) + i18n.Translate("fl", lang),
					"title": i18n.Translate("rt", lang),
					"url":   urlForPath(cPath),
				})
				return
			}

			destDir := filepath.Join(conf.GoFile, cPath)
			if !utils.IsPathSafe(destDir) {
				c.Status(http.StatusForbidden)
				return
			}

			dest := filepath.Join(destDir, filepath.Base(file.Filename))
			stat := i18n.Translate("sc", lang)
			if err := c.SaveUploadedFile(file, dest); err != nil {
				stat = i18n.Translate("fl", lang)
			} else {
				logAction(c, "上传", filepath.Join(cPath, filepath.Base(file.Filename)))
			}
			renderHTML(c, "msg.tmpl", gin.H{
				"msg":   i18n.Translate("upFile", lang) + stat,
				"title": i18n.Translate("rt", lang),
				"url":   urlForPath(cPath),
			})
		})

		// 分片上传 - 查询已上传分片
		r.POST("/do/chunk/check", func(c *gin.Context) {
			fileId := c.PostForm("fileId")
			dir := chunkDir(fileId)
			if dir == "" {
				c.JSON(http.StatusOK, gin.H{"uploaded": []int{}})
				return
			}
			totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))
			uploaded := []int{}
			for i := 0; i < totalChunks; i++ {
				if utils.Exist(filepath.Join(dir, strconv.Itoa(i))) {
					uploaded = append(uploaded, i)
				}
			}
			c.JSON(http.StatusOK, gin.H{"uploaded": uploaded})
		})

		// 分片上传 - 上传单个分片
		r.POST("/do/chunk/upload", func(c *gin.Context) {
			fileId := c.PostForm("fileId")
			dir := chunkDir(fileId)
			if dir == "" {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			chunkIndex, err := strconv.Atoi(c.PostForm("chunkIndex"))
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			if err := c.SaveUploadedFile(file, filepath.Join(dir, strconv.Itoa(chunkIndex))); err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			c.JSON(http.StatusOK, gin.H{"stat": true})
		})

		// 分片上传 - 合并分片为最终文件
		r.POST("/do/chunk/merge", func(c *gin.Context) {
			fileId := c.PostForm("fileId")
			dir := chunkDir(fileId)
			if dir == "" {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			totalChunks, err := strconv.Atoi(c.PostForm("totalChunks"))
			if err != nil || totalChunks <= 0 {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			destDir := filepath.Join(conf.GoFile, c.PostForm("path"))
			if !utils.IsPathSafe(destDir) {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			destPath := filepath.Join(destDir, filepath.Base(c.PostForm("fileName")))
			out, err := os.Create(destPath)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			defer func() {
				out.Close()
				os.RemoveAll(dir)
			}()
			for i := 0; i < totalChunks; i++ {
				chunk, err := os.Open(filepath.Join(dir, strconv.Itoa(i)))
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"stat": false})
					return
				}
				_, copyErr := io.Copy(out, chunk)
				chunk.Close()
				if copyErr != nil {
					c.JSON(http.StatusOK, gin.H{"stat": false})
					return
				}
			}
			logAction(c, "上传", filepath.Join(c.PostForm("path"), filepath.Base(c.PostForm("fileName"))))
			c.JSON(http.StatusOK, gin.H{"stat": true})
		})

		// API 上传（返回 JSON，适合脚本调用）
		r.POST("/api/upload", func(c *gin.Context) {
			destDir := filepath.Join(conf.GoFile, c.PostForm("path"))
			if !utils.IsPathSafe(destDir) {
				c.JSON(http.StatusForbidden, gin.H{"stat": false, "msg": "forbidden"})
				return
			}
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"stat": false, "msg": "no file"})
				return
			}
			dest := filepath.Join(destDir, filepath.Base(file.Filename))
			if err := c.SaveUploadedFile(file, dest); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"stat": false, "msg": err.Error()})
				return
			}
			logAction(c, "上传", filepath.Join(c.PostForm("path"), filepath.Base(file.Filename)))
			c.JSON(http.StatusOK, gin.H{"stat": true})
		})

	}

	// 编辑路由（仅普通模式）
	if !reader {
		// 新建文件
		r.POST("/do/newfile", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			filePath := filepath.Join(conf.GoFile, c.PostForm("path"), filepath.Base(c.PostForm("filename")))
			ok := false
			if utils.IsPathSafe(filePath) && !utils.Exist(filePath) {
				f, err := os.Create(filePath)
				if err == nil {
					f.Close()
					ok = true
					logAction(c, "新建", filepath.Join(c.PostForm("path"), filepath.Base(c.PostForm("filename"))))
				}
			}
			c.JSON(http.StatusOK, gin.H{"stat": ok})
		})

		// 新建文件夹
		r.POST("/do/newdir", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			dirPath := filepath.Join(conf.GoFile, c.PostForm("path"), filepath.Base(c.PostForm("dirname")))
			ok := false
			if utils.IsPathSafe(dirPath) {
				if err := os.Mkdir(dirPath, 0755); err == nil {
					ok = true
					logAction(c, "新建", filepath.Join(c.PostForm("path"), filepath.Base(c.PostForm("dirname")))+"/")
				}
			}
			c.JSON(http.StatusOK, gin.H{"stat": ok})
		})

		// 解压文件
		r.POST("/do/unzip", func(c *gin.Context) {
			relPath := c.PostForm("path")
			absPath := filepath.Join(conf.GoFile, relPath)
			if !utils.IsPathSafe(absPath) {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			ext := strings.ToLower(filepath.Ext(relPath))
			ok := false
			switch ext {
			case ".zip":
				ok = utils.Unzip(relPath)
			case ".gz":
				cmd := exec.Command("tar", "-zxvf", absPath, "-C", filepath.Dir(absPath))
				ok = cmd.Run() == nil
			}
			c.JSON(http.StatusOK, gin.H{"stat": ok})
		})

		// 保存代码
		r.POST("/do/save/", func(c *gin.Context) {
			filePath := filepath.Join(conf.GoFile, c.PostForm("path"))
			if !utils.IsPathSafe(filePath) {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			defer file.Close()
			data := c.PostForm("data")
			if len(data) > 0 && data[len(data)-1] == '\n' {
				data = data[:len(data)-1] // 去掉新行
			}
			_, err = file.WriteString(data)
			if err == nil {
				logAction(c, "保存", c.PostForm("path"))
			}
			c.JSON(http.StatusOK, gin.H{"stat": err == nil})
		})

		// 编辑代码
		r.POST("/edite/", func(c *gin.Context) {
			filePath := filepath.Join(conf.GoFile, c.PostForm("path"))
			if !utils.IsPathSafe(filePath) {
				c.Status(http.StatusForbidden)
				return
			}
			logAction(c, "编辑", c.PostForm("path"))
			file, err := os.Open(filePath)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer file.Close()
			data, _ := io.ReadAll(file)
			renderHTML(c, "editor.tmpl", gin.H{
				"data": strings.TrimSpace(string(data)),
				"path": c.PostForm("path"),
			})
		})

		// 删除文件/文件夹
		r.POST("/do/rm", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			path := filepath.Join(conf.GoFile, c.PostForm("path"))
			if !utils.IsPathSafe(path) {
				c.JSON(http.StatusOK, gin.H{"stat": false})
				return
			}
			err := os.RemoveAll(path)
			c.JSON(http.StatusOK, gin.H{"stat": err == nil})
		})
	}

	r.Run("0.0.0.0:" + conf.GoFilePort)
}

func init() {
	flag.StringVar(&conf.GoFile, "path", "./", "goFile path")
	flag.StringVar(&conf.GoFilePort, "port", "8089", "goFile web port")
	readerPtr   := flag.Bool("r",  false, "Read-only mode")
	uploaderPtr := flag.Bool("ru", false, "Read-only + allow upload")
	flag.Parse()
	if *readerPtr {
		reader = true
	}
	if *uploaderPtr {
		reader   = true
		uploader = true
	}
}

func localIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.To4() == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

func main() {
	if conf.GoFile == "./" {
		cwd, err := os.Getwd()
		if err == nil {
			conf.GoFile = cwd
		}
	}
	conf.GoFile = filepath.Clean(conf.GoFile) + string(filepath.Separator)
	fmt.Println(strings.Repeat("-", 44))
	fmt.Println("Directory : " + conf.GoFile)
	fmt.Println("Port      : " + conf.GoFilePort)
	for _, ip := range localIPs() {
		fmt.Println("Access    : http://" + ip + ":" + conf.GoFilePort)
	}
	fmt.Println(strings.Repeat("-", 44))
	web()
}
