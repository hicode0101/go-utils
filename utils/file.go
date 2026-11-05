package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// utilFile 文件与目录工具。
//
// 路径处理统一使用 filepath（而非 path），保证 Windows 反斜杠路径与
// Unix 正斜杠路径都能正确处理。
type utilFile struct{}

/* ---------------- 打开与创建 ---------------- */

// Create 创建或截断文件（已存在则清空内容），返回可写文件句柄。
// 场景：准备一个新日志/导出文件。用完记得 f.Close()。
func (f *utilFile) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// Open 以只读方式打开文件。场景：逐行读取大文件（配合 bufio.Scanner）。
func (f *utilFile) Open(name string) (*os.File, error) {
	return os.Open(name)
}

// OpenFile 按指定 flag 与权限打开文件。
// 如追加写入：OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)。
func (f *utilFile) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

/* ---------------- 读写 ---------------- */

// Read 一次性读取整个文件的字节内容。场景：读小配置文件；大文件请用 Open + bufio。
func (f *utilFile) Read(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// ReadText 一次性读取整个文件为字符串。场景：读取模板、JSON 配置。
func (f *utilFile) ReadText(name string) (string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write 写入字节内容（覆盖已存在文件），文件不存在则创建（0644）。
// 场景：落盘计算结果。
func (f *utilFile) Write(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

// WriteText 写入字符串（覆盖已存在文件）。场景：导出文本报告。
func (f *utilFile) WriteText(name, content string) error {
	return os.WriteFile(name, []byte(content), 0644)
}

// Append 追加字节内容，文件不存在则创建。
// 场景：持续追加的采集数据、抓包记录。
func (f *utilFile) Append(name string, data []byte) error {
	handle, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer handle.Close()
	_, err = handle.Write(data)
	return err
}

// AppendText 追加一行文本（自动补换行）。场景：写 CSV、简易日志。
func (f *utilFile) AppendText(name, line string) error {
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	return f.Append(name, []byte(line))
}

// Copy 复制文件，覆盖已存在的目标文件并保留源文件权限。
// 场景：备份配置、构建产物搬运。
func (f *utilFile) Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	// 保留源文件权限
	if info, statErr := in.Stat(); statErr == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return nil
}

/* ---------------- 存在性与信息 ---------------- */

// Exists 判断文件或目录是否存在。
// 场景：启动时检查配置文件是否就位。
func (f *utilFile) Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// Stat 取文件信息（大小、修改时间、权限等）。
// 场景：判断文件是否被更新（比较 ModTime()）。
func (f *utilFile) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Size 返回文件字节数，文件不存在时返回 -1。
// 场景：上传前校验大小、FormatBytes(Size(path)) 展示。
func (f *utilFile) Size(name string) int64 {
	info, err := os.Stat(name)
	if err != nil || info.IsDir() {
		return -1
	}
	return info.Size()
}

/* ---------------- 目录操作 ---------------- */

// MkdirAll 递归创建目录（父目录不存在一并创建），已存在时为空操作。
// 场景：启动时确保 logs/、data/ 目录存在。
func (f *utilFile) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// ListFiles 递归列出目录下所有文件的完整路径；ext 非空时按扩展名过滤
// （如 ".go"，大小写不敏感），ext 为空串时返回全部文件。
// 场景：扫描指定目录下的所有 .csv 文件做批量处理。
func (f *utilFile) ListFiles(dir, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if ext == "" || strings.EqualFold(filepath.Ext(p), ext) {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

/* ---------------- 删除与重命名 ---------------- */

// Rename 重命名/移动文件或目录。场景：临时文件写完后原子改名为正式文件。
func (f *utilFile) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Remove 删除文件或空目录。场景：清理单个临时文件。
func (f *utilFile) Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll 递归删除目录及其全部内容。场景：清理缓存目录。
// 注意：不可逆操作，调用前务必确认路径来源可信（避免拼接用户输入）。
func (f *utilFile) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

/* ---------------- 路径处理 ---------------- */

// Join 拼接路径元素（自动使用平台分隔符，并做规范化）。
// Join("logs", "2024", "app.log") = "logs/2024/app.log"（Unix）。
func (f *utilFile) Join(elems ...string) string {
	return filepath.Join(elems...)
}

// Dir 取路径的目录部分。Dir("/app/src/main.go") = "/app/src"。
func (f *utilFile) Dir(fullPath string) string {
	return filepath.Dir(fullPath)
}

// Name 取路径的文件名（含扩展名）。Name("/app/main.go") = "main.go"。
func (f *utilFile) Name(fullPath string) string {
	return filepath.Base(fullPath)
}

// NameWithoutExt 取不含扩展名的文件名。
// NameWithoutExt("/app/main.go") = "main"。场景：导出文件派生命名。
func (f *utilFile) NameWithoutExt(fullPath string) string {
	base := filepath.Base(fullPath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// Ext 取扩展名（含点）。Ext("/app/main.go") = ".go"。
func (f *utilFile) Ext(fullPath string) string {
	return filepath.Ext(fullPath)
}

/* ---------------- 程序路径 ---------------- */

// WorkDir 返回当前工作目录。场景：拼接相对路径的资源位置。
func (f *utilFile) WorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// ExePath 返回当前可执行文件的绝对路径。
// 场景：以可执行文件位置为基准定位配置文件（不受启动目录影响）。
func (f *utilFile) ExePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return exe
}

func (_self *utilFile) GetCurrentDir() string {

	workPath, _ := os.Getwd()
	return workPath
}

func (_self *utilFile) GetCurrentExe() string {

	fullPath, _ := os.Executable()
	return fullPath
}
