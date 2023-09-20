/*
@Time : 2021/1/9 18:26
@Author : LiuKun
@File : file
@Software: GoLand
@Description:
*/

package fileutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// ReplacePathExt 替换后缀名
func ReplacePathExt(path, newExt string) string {
	return strings.ReplaceAll(path, filepath.Ext(path), newExt)
}

// IsDir 是否是目录
func IsDir(name string) bool {
	if info, err := os.Stat(name); err == nil {
		return info.IsDir()
	}
	return false
}

// FileIsExisted 文件是否存在
func FileIsExisted(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// CreateFile 创建文件, 需要Close
func CreateFile(Path string) (*os.File, error) {

	dir, _ := filepath.Split(Path)
	err := MakeDir(dir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("create file error :%s", err.Error()))
	}

	f, err := os.Create(Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// MakeDir 创建文件夹
func MakeDir(dir string) error {
	if !FileIsExisted(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil { //os.ModePerm
			return errors.New(fmt.Sprintf("make dir failed :%s", err.Error()))
		}
	}
	return nil
}

// CopyFileByIoCopy 使用io.Copy文件, 复制文件过程中一定要注意将原始文件的权限也要复制过去，否则可能会导致可执行文件不能执行等问题
func CopyFileByIoCopy(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	//desFile, err := os.Create(des)  //无法复制源文件的所有权限
	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) //复制源文件的所有权限
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

// CopyFileByIoUtil 使用ioutil.WriteFile()和ioutil.ReadFile() 复制文件
func CopyFileByIoUtil(src, des string) (written int64, err error) {
	//获取源文件的权限
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	fi, _ := srcFile.Stat()
	perm := fi.Mode()
	srcFile.Close()

	input, err := os.ReadFile(src)
	if err != nil {
		return 0, err
	}

	err = os.WriteFile(des, input, perm)
	if err != nil {
		return 0, err
	}

	return int64(len(input)), nil
}

// CopyFileByOs 使用os.Read()和os.Write()复制文件
func CopyFileByOs(src, des string, bufSize int) (written int64, err error) {
	if bufSize <= 0 {
		bufSize = 1 * 1024 * 1024 //1M
	}
	buf := make([]byte, bufSize)

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	desFile, err := os.OpenFile(des, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	count := 0
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}

		if n == 0 {
			break
		}

		if wn, err := desFile.Write(buf[:n]); err != nil {
			return 0, err
		} else {
			count += wn
		}
	}

	return int64(count), nil
}

// CopyDir 复制整个文件夹
func CopyDir(srcPath, desPath string) error {
	//检查目录是否正确
	if srcInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else {
		if !srcInfo.IsDir() {
			return errors.New("源路径不是一个正确的目录！")
		}
	}

	if desInfo, err := os.Stat(desPath); err != nil {
		return err
	} else {
		if !desInfo.IsDir() {
			return errors.New("目标路径不是一个正确的目录！")
		}
	}

	if strings.TrimSpace(srcPath) == strings.TrimSpace(desPath) {
		return errors.New("源路径与目标路径不能相同！")
	}

	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		//复制目录是将源目录中的子目录复制到目标路径中，不包含源目录本身
		if path == srcPath {
			return nil
		}

		//生成新路径
		destNewPath := strings.Replace(path, srcPath, desPath, -1)

		if !f.IsDir() {
			_, _ = CopyFileByIoCopy(path, destNewPath)
		} else {
			if !FileIsExisted(destNewPath) {
				return MakeDir(destNewPath)
			}
		}

		return nil
	})

	return err
}

// ListDir 遍历指定文件夹中的所有文件（不进入下一级子目录）
// 获取指定路径下的所有文件及文件夹，只搜索当前路径，不进入下一级目录，可匹配后缀过滤（suffix为空则不过滤）
func ListDir(dir, suffix string) (files []string, err error) {
	files = []string{}

	_dir, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	suffix = strings.ToLower(suffix) //匹配后缀

	for _, _file := range _dir {

		if len(suffix) == 0 || strings.HasSuffix(strings.ToLower(_file.Name()), suffix) {
			//文件后缀匹配
			files = append(files, path.Join(dir, _file.Name()))
		}
	}

	return files, nil
}

// WalkDir 遍历指定路径及其子目录中的所有文件
// 获取指定路径下以及所有子目录下的所有文件，可匹配后缀过滤（suffix为空则不过滤）
func WalkDir(dir, suffix string) (files []string, err error) {
	files = []string{}

	err = filepath.Walk(dir, func(fName string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			//忽略目录
			return nil
		}

		if len(suffix) == 0 || strings.HasSuffix(strings.ToLower(fi.Name()), suffix) {
			//文件后缀匹配
			files = append(files, fName)
		}

		return nil
	})

	return files, err
}

// RenameFile 重命名文件
func RenameFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// RemoveFile 删除文件
func RemoveFile(filename string) error {
	return os.Remove(filename)
}

// RemoveAll 删除文件夹及其包含的所有子目录和所有文件
func RemoveAll(dir string) error {
	return os.RemoveAll(dir)
}

// ReadFile 读取文件
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// GetFileModTime 获取文件修改时间
func GetFileModTime(filename string) (time.Time, error) {

	fi, err := os.Stat(filename)
	if err != nil {
		return time.Now(), errors.New(fmt.Sprintf("stat file info error :%s", err.Error()))
	}

	return fi.ModTime(), nil
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {

	fi, err := os.Stat(filename)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("stat file info error :%s", err.Error()))
	}

	return fi.Size(), nil
}

// WriteDataToFile 写数据到文件
func WriteDataToFile(filePath string, data interface{}) error {

	f, err := CreateFile(filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("create file fail: %s\n", err.Error()))
	}
	defer f.Close()

	//写入文件
	var v []byte

	switch data.(type) {
	case []byte:
		v = data.([]byte)
	case string:
		v = []byte(data.(string))
	default:
		jsonByte, err := json.Marshal(data)
		if err != nil {
			return errors.New(fmt.Sprintf("convert data to json fail: %s", err.Error()))
		}
		v = jsonByte
	}
	_, e := f.Write(v)
	if e != nil {
		return errors.New(fmt.Sprintf("save data fail: %s", e.Error()))
	}

	return nil
}

// CheckMultiFileExisted 检查多个文件是否存在且修改时间大于interval
func CheckMultiFileExisted(timeout, interval time.Duration, filePaths ...string) (string, error) {

	fs := make([]string, 0)
	for _, f := range filePaths {
		if len(f) > 0 {
			fs = append(fs, f)
		}
	}
	if len(fs) < 1 {
		return "", errors.New("no valid file path")
	}

	t := time.Now()
	for {
		time.Sleep(time.Second / 10)
		if time.Now().Sub(t) > timeout {
			return "", errors.New("check time out")
		}

		for _, f := range filePaths {
			if FileIsExisted(f) {
				mt, _ := GetFileModTime(f)
				if time.Now().Sub(mt) > interval {
					return f, nil
				}
			}
		}
	}

}

// CheckFileExisted 检测单个文件是否存在
func CheckFileExisted(filePath string, timeout time.Duration) error {
	_, e := CheckMultiFileExisted(timeout, 2*time.Second, filePath)
	return e
}
