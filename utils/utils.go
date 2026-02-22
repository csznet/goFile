package utils

import (
	"archive/zip"
	"goFile/conf"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// IsPathSafe checks that path is within the GoFile root directory,
// preventing path traversal attacks.
func IsPathSafe(path string) bool {
	root := filepath.Clean(conf.GoFile)
	cleaned := filepath.Clean(path)
	return cleaned == root || strings.HasPrefix(cleaned, root+string(filepath.Separator))
}

// Unzip extracts a zip archive. src is a path relative to conf.GoFile.
// Each entry's destination is validated to prevent Zip Slip attacks.
func Unzip(src string) bool {
	OutPath := pathOutConv(src)
	fullSrc := filepath.Join(conf.GoFile, src)
	fr, err := zip.OpenReader(fullSrc)
	if err != nil {
		return false
	}
	defer fr.Close()

	for _, file := range fr.Reader.File {
		destPath := filepath.Join(OutPath, file.Name)
		// Zip Slip protection: reject entries that escape the output directory
		if !IsPathSafe(destPath) {
			return false
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return false
			}
			continue
		}
		// Ensure parent directory exists for file entries without explicit dir entries
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return false
		}
		r, err := file.Open()
		if err != nil {
			return false
		}
		newFile, err := os.Create(destPath)
		if err != nil {
			r.Close()
			return false
		}
		io.Copy(newFile, r)
		newFile.Close()
		r.Close()
	}
	return true
}

// pathOutConv returns the directory that contains the given file path.
func pathOutConv(src string) string {
	return filepath.Dir(filepath.Join(conf.GoFile, src)) + string(filepath.Separator)
}

// Exist reports whether path exists on the filesystem.
func Exist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// GetFiles returns directory and file listings for the given glob pattern.
func GetFiles(path string) conf.Info {
	getFile, err := filepath.Glob(path)
	if err != nil {
		return conf.Info{}
	}

	var info conf.Info
	ZipList := []string{"zip", "gz"}

	for _, im := range getFile {
		if !Exist(im) {
			continue
		}
		s, err := os.Stat(im)
		if err != nil {
			continue
		}
		relPath := getRelativePath(im)
		if s.IsDir() {
			info.Dirs = append(info.Dirs, conf.Dir{
				DirPath: relPath,
				DirName: filepath.Base(im),
			})
		} else {
			info.Files = append(info.Files, conf.File{
				FilePath: relPath,
				FileName: filepath.Base(im),
				IsZip:    isZipFile(im, ZipList),
			})
		}
	}
	return info
}

func getRelativePath(im string) string {
	return filepath.ToSlash(strings.TrimPrefix(im, conf.GoFile))
}

func isZipFile(fileName string, zipList []string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, zipExt := range zipList {
		if "."+zipExt == ext {
			return true
		}
	}
	return false
}

func GetPrevPath(cPath string) string {
	if cPath == "" || cPath == "/" {
		return "/"
	}
	prevPath := path.Dir(cPath)
	if prevPath == "." || prevPath == "/" {
		return "/"
	}
	return "/d/" + prevPath
}
