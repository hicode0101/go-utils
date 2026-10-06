package utils

import (
	"os"
	"path"
)

type utilFile struct {
}

func (_self *utilFile) FileCreate(name string) (*os.File, error) {

	return os.Create(name)

}

func (_self *utilFile) FileOpen(name string) (*os.File, error) {

	//文件读取可以使用 bufio、ioutil 库

	return os.Open(name)

}

func (_self *utilFile) FileOpenMod(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (_self *utilFile) FileRead(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (_self *utilFile) MakeDir(path string, perm os.FileMode) error {
	//return os.Mkdir(path, perm)
	return os.MkdirAll(path, perm)
}

// 文件重命名
func (_self *utilFile) FileReName(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// 删除文件
func (_self *utilFile) FileDel(name string) error {
	return os.Remove(name)
}

// 删除整个目录
func (_self *utilFile) DirDel(path string) error {
	return os.RemoveAll(path)
}

// 取文件的信息
func (_self *utilFile) FileInfo(name string) (os.FileInfo, error) {
	//其他文件信息，可以通过FileOpen方法拿到os.File
	return os.Stat(name)

}

// 判断文件/或文件夹是否存在
func (_self *utilFile) FileIsExist(name string) bool {
	var exist = true
	if _, err := os.Stat(name); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 取文件所在的目录 例如：
// services.FileDir("d://app//src//main.go") = d:/app/src
func (_self *utilFile) FileDir(fullPath string) string {
	return path.Dir(fullPath)
}

// 取文件名，包含扩展名 例如：
// services.FileFullName("d://main.go") = main.go
func (_self *utilFile) FileFullName(fullPath string) string {
	return path.Base(fullPath)
}

// 取文件扩展名 例如：
// services.FileExt("d://main.go") = .go
func (_self *utilFile) FileExt(fullPath string) string {
	return path.Ext(fullPath)
}

func (_self *utilFile) GetCurrentDir() string {

	workPath, _ := os.Getwd()
	return workPath
}

func (_self *utilFile) GetCurrentExe() string {

	fullPath, _ := os.Executable()
	return fullPath
}
