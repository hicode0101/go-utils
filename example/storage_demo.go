package main

import (
	"fmt"
	"os"

	"github.com/hicode0101/go-utils/utils"
)

// storageDemo 文件与压缩：File / Zip。
func storageDemo() {
	banner("File 文件与目录")
	dir, err := os.MkdirTemp("", "go-utils-example-*")
	if err != nil {
		fmt.Println("创建临时目录失败:", err)
		return
	}
	defer os.RemoveAll(dir) // 示例结束清理

	p := utils.File.Join(dir, "demo.txt") // 跨平台路径拼接
	if err := utils.File.WriteText(p, "hello go-utils\n"); err == nil {
		fmt.Println(`WriteText ->`, p)
	}
	_ = utils.File.AppendText(p, "第二行内容")
	if content, err := utils.File.ReadText(p); err == nil {
		fmt.Printf("ReadText -> %q\n", content)
	}
	fmt.Println("Exists =", utils.File.Exists(p), "| Size =", utils.File.Size(p), "字节")

	copyPath := utils.File.Join(dir, "demo-copy.txt")
	if err := utils.File.Copy(p, copyPath); err == nil {
		fmt.Println("Copy ->", copyPath, "| Exists =", utils.File.Exists(copyPath))
	}
	fmt.Println("ListFiles(dir, \"\") =", mustList(utils.File.ListFiles(dir, "")))

	// 路径处理（不访问磁盘）
	fmt.Printf(`Name("/a/b/c.txt")=%s Dir=%s Ext=%s NameWithoutExt=%s`+"\n",
		utils.File.Name("/a/b/c.txt"), utils.File.Dir("/a/b/c.txt"),
		utils.File.Ext("/a/b/c.txt"), utils.File.NameWithoutExt("/a/b/c.txt"))
	fmt.Println("WorkDir() =", utils.File.WorkDir())
	// 其它方法：Create / Open / OpenFile / Read / Write / Append / MkdirAll / Rename / Remove / RemoveAll / Stat / ExePath

	banner("Zip 压缩解压")
	zipPath := utils.File.Join(dir, "demo.zip")
	if err := utils.Zip.ZipEntryBytes(zipPath, "docs/readme.txt", []byte("zip 里的内容")); err == nil {
		fmt.Println(`ZipEntryBytes(zip, "docs/readme.txt", ...) 完成`)
	}
	if data, err := utils.Zip.ReadEntryBytes(zipPath, "docs/readme.txt"); err == nil {
		fmt.Printf("ReadEntryBytes -> %q\n", string(data))
	}
	if names, err := utils.Zip.ListFileNames(zipPath, "docs"); err == nil {
		fmt.Println(`ListFileNames(zip, "docs") =`, names, "（docs 层级文件名，不含更深子目录）")
	}

	zipPath2 := utils.File.Join(dir, "files.zip")
	if err := utils.Zip.ZipFiles(zipPath2, p, copyPath); err == nil {
		outDir := utils.File.Join(dir, "unzipped")
		if err := utils.Zip.Unzip(zipPath2, outDir); err == nil {
			fmt.Println("ZipFiles + Unzip 往返 ->", mustList(utils.File.ListFiles(outDir, "")))
		}
	}
	if gz, err := utils.Zip.Gzip([]byte("gzip compress me")); err == nil {
		back, _ := utils.Zip.UnGzip(gz)
		fmt.Printf("Gzip/UnGzip 往返: %d 字节 -> %q\n", len(gz), string(back))
	}
	if s, err := utils.Zip.GzipStr("字符串压缩"); err == nil {
		back, _ := utils.Zip.UnGzipStr(s)
		fmt.Printf("GzipStr/UnGzipStr 往返 -> %q\n", back)
	}
	// 其它方法：ZipDir(目录打包) / Zlib / UnZlib
}

func mustList(names []string, err error) []string {
	if err != nil {
		return nil
	}
	return names
}
