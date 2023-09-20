/*
@Time : 2022/6/29 16:20
@Author : LiuKun
@File : zip
@Software: GoLand
@Description:
*/

package fileutil

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
)

// ZipDeCompress 解压
func ZipDeCompress(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		filename := dest + file.Name
		w, err := CreateFile(filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()
	}
	return nil
}

// ZipCompress dest 为zip目录， src， otherSrc...为压缩文件或文件夹
func ZipCompress(dest string, src ...string) error {

	if len(src) < 1 {
		return errors.New("没有可压缩的文件")
	}

	files := make([]*os.File, 0)
	for _, s := range src {
		f, err := os.Open(s)
		if err != nil {
			return err
		}
		files = append(files, f)
	}
	d, err := CreateFile(dest)
	if err != nil {
		return err
	}
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compressZip(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func compressZip(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		// 增加对空目录的判断
		if len(fileInfos) <= 0 {
			header, err := zip.FileInfoHeader(info)
			header.Name = prefix
			if err != nil {
				fmt.Println("error is:" + err.Error())
				return err
			}
			_, err = zw.CreateHeader(header)
			if err != nil {
				fmt.Println("create error is:" + err.Error())
				return err
			}
			file.Close()
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compressZip(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
