package utils

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// utilZip 压缩工具。
//
// zip 打包/解包方法对照 hutool ZipUtil，gzip/zlib 方法对照其 Gzip/Zlib 系列；
// 在 hutool 语义基础上补充了 zip-slip 路径穿越防护与 zip 炸弹（解压尺寸上限）防护。
// zip 条目名统一使用正斜杠 / 分隔（zip 规范），对外路径参数使用 filepath 处理，
// 保证 Windows 反斜杠路径与 Unix 正斜杠路径都能正确工作。
type utilZip struct{}

// zipMaxUnzipSize 解压尺寸上限（2GB）。
// 对应 hutool unzip(zipFile, outFile, limit) 的 limit 预检与 getStream 的
// LimitedInputStream：条目声明的解压后总尺寸超过该值即判定为 zip 炸弹并拒绝解压，
// 防止恶意压缩包耗尽磁盘与内存。
const zipMaxUnzipSize int64 = 2 << 30

/* ---------------- zip 打包 ---------------- */

// ZipDir 把整个目录压缩为 zip 文件，递归包含全部子目录与文件。
// withRoot 为 true 时压缩包内包含 srcDir 本身这一层目录（包内路径以 srcDir 的
// 基本名开头），为 false 时仅包含目录下的内容；对应 hutool zip(srcPath, zipPath, withSrcDir)。
// 场景：打包日志目录、导出数据目录供下载。
// 注意：zipPath 不能位于 srcDir 内部（防止边压缩边把输出写进输出，对应 hutool validateFiles 校验）。
func (z *utilZip) ZipDir(srcDir, zipPath string, withRoot bool) error {
	info, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("源路径 [%s] 不是目录，压缩单个文件或混合打包请用 ZipFiles", srcDir)
	}
	return zipPack(zipPath, []string{srcDir}, withRoot)
}

// ZipFiles 把多个文件或目录压入同一个 zip 文件（文件不存在或目标非法时返回 error）。
// 每个路径以自身基本名出现在压缩包根目录，目录会连同其内部结构一起打包
// （对照 hutool append 的语义：新文件添加在 zip 根目录，文件夹包括其本身和内容）。
// 场景：把若干日志、配置文件打包归档或上传。
func (z *utilZip) ZipFiles(zipPath string, filePaths ...string) error {
	if len(filePaths) == 0 {
		return errors.New("至少需要提供一个待压缩路径")
	}
	return zipPack(zipPath, filePaths, true)
}

// ZipEntryBytes 把一段字节作为单个条目写入 zip 文件。
// zip 不存在时新建仅含该条目的压缩包；zip 已存在时读出旧条目重写并追加新条目
// （标准库不支持原地追加，对照 hutool append 借助 NIO FileSystem 重写压缩包的实现）。
// 若包内已存在同名条目，保留原条目、跳过写入（对照 hutool append 默认不覆盖）。
// 场景：向归档包中追加生成的清单、报表文件。
func (z *utilZip) ZipEntryBytes(zipPath, entryPath string, data []byte) error {
	entryPath = zipNormalizeEntryName(entryPath)
	if entryPath == "" {
		return errors.New("zip 条目路径 entryPath 不能为空")
	}
	newEntry := zipEntryData{name: entryPath, mod: time.Now(), data: data}
	if _, err := os.Stat(zipPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// zip 不存在：新建仅含该条目的压缩包
		return zipWriteAll(zipPath, []zipEntryData{newEntry})
	}
	oldEntries, err := zipReadAllEntries(zipPath)
	if err != nil {
		return err
	}
	all := make([]zipEntryData, 0, len(oldEntries)+1)
	exists := false
	for _, e := range oldEntries {
		if e.name == entryPath {
			exists = true
			// 同名条目保留旧值（对照 hutool append 默认不覆盖），不可跳过
		}
		all = append(all, e)
	}
	if !exists {
		all = append(all, newEntry)
	}
	return zipWriteAll(zipPath, all)
}

/* ---------------- zip 解包与读取 ---------------- */

// Unzip 把 zip 文件解压到指定目录（目录不存在会自动创建）。
// 安全防护：
//   - zip-slip：条目名含 ".." 或以 "/"、"\" 开头时返回 error，拒绝路径穿越；
//   - zip 炸弹：解压总尺寸超过 2GB 时返回 error（对应 hutool unzip 的 limit 限制，
//     先按条目声明尺寸预检，解压过程中再按实际写入量二次校验，防止声明尺寸造假）。
//
// 场景：还原上传的压缩包、解压发布产物。
func (z *utilZip) Unzip(zipPath, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 第一遍预检：zip-slip 校验 + 条目声明尺寸累计（对应 hutool unzip limit 预检）
	var declared int64
	for _, f := range r.File {
		if zipIsUnsafeName(f.Name) {
			return fmt.Errorf("zip 条目 [%s] 含不安全路径（.. 或绝对路径），已拒绝解压（zip-slip 防护）", f.Name)
		}
		declared += int64(f.UncompressedSize64)
		if declared > zipMaxUnzipSize {
			return errors.New("zip 声明的解压总尺寸超过 2GB 上限，疑似 zip 炸弹，已拒绝解压")
		}
	}

	outAbs, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	var actual int64
	for _, f := range r.File {
		name := f.Name
		if runtime.GOOS == "windows" {
			// Windows 文件名不允许出现 *，替换为 _（对照 hutool gitee issue #I4ZDQI）
			name = strings.ReplaceAll(name, "*", "_")
		}
		dest := filepath.Join(outDir, filepath.FromSlash(name))
		// 双重保险：最终落盘路径必须仍在解压目录内
		if absDest, absErr := filepath.Abs(dest); absErr != nil {
			return absErr
		} else if !zipIsSubPath(absDest, outAbs) {
			return fmt.Errorf("zip 条目 [%s] 解压路径越界，已拒绝", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		perm := f.Mode().Perm()
		if perm == 0 {
			perm = 0644
		}
		out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			rc.Close()
			return err
		}
		n, copyErr := io.Copy(out, rc)
		actual += n
		out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
		// 实际写入量二次校验：条目头中声明的尺寸可能被伪造
		if actual > zipMaxUnzipSize {
			return errors.New("zip 实际解压总尺寸超过 2GB 上限，疑似 zip 炸弹，已中止")
		}
	}
	return nil
}

// ReadEntryBytes 读取 zip 内单个条目的内容字节，条目不存在时返回 error。
// entryPath 为压缩包内的完整路径（如 "logs/app.log"，Windows 反斜杠会自动转换）。
// 场景：不落盘直接读取压缩包内的配置、清单文件。
// 单条目解压尺寸超过 2GB 时返回 error（对应 hutool getStream 的 LimitedInputStream 防护）。
func (z *utilZip) ReadEntryBytes(zipPath, entryPath string) ([]byte, error) {
	entryPath = zipNormalizeEntryName(entryPath)
	if entryPath == "" {
		return nil, errors.New("zip 条目路径 entryPath 不能为空")
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	for _, f := range r.File {
		if zipNormalizeEntryName(f.Name) != entryPath {
			continue
		}
		if f.FileInfo().IsDir() {
			return nil, fmt.Errorf("zip 条目 [%s] 是目录，无内容可读", entryPath)
		}
		if int64(f.UncompressedSize64) > zipMaxUnzipSize {
			return nil, fmt.Errorf("zip 条目 [%s] 解压后尺寸超过 2GB 上限，疑似 zip 炸弹", entryPath)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("zip 文件 [%s] 中不存在条目 [%s]", zipPath, entryPath)
}

// ListFileNames 列出 zip 内 dir 前缀目录下的文件名（只显示文件，不显示目录），
// 返回去掉 dir 前缀后的名字，对照 hutool listFileNames。
// dir 为空串（或纯空白）时列出压缩包根目录下的全部文件名；dir 结尾的 "/" 可省略。
// 场景：解压前先预览压缩包内有哪些文件。
func (z *utilZip) ListFileNames(zipPath, dir string) ([]string, error) {
	if strings.TrimSpace(dir) != "" {
		// 目录尾部补 "/"
		if !strings.HasSuffix(dir, "/") {
			dir += "/"
		}
	} else {
		dir = ""
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	names := make([]string, 0)
	for _, f := range r.File {
		name := f.Name
		if dir == "" || strings.HasPrefix(name, dir) {
			suffix := strings.TrimPrefix(name, dir)
			// 剩余部分不含 / 才是当前层级的文件（目录与子目录内容均被排除）
			if suffix != "" && !strings.Contains(suffix, "/") {
				names = append(names, suffix)
			}
		}
	}
	return names, nil
}

/* ---------------- gzip ---------------- */

// Gzip 使用 gzip 算法压缩字节流（默认压缩级别，对应 hutool gzip）。
// 场景：压缩接口响应体、日志归档前预处理。
func (z *utilZip) Gzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnGzip 解压 gzip 压缩的字节流（对应 hutool unGzip）。
// 数据不是合法 gzip 格式时返回 error。场景：还原 Gzip 压缩的响应或文件内容。
func (z *utilZip) UnGzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// GzipStr 字符串便捷版：压缩字符串，结果以字符串形式承载压缩后的字节。
// 场景：Base64/Hex 编码前先压缩以缩短传输内容。
func (z *utilZip) GzipStr(s string) (string, error) {
	out, err := z.Gzip([]byte(s))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// UnGzipStr 字符串便捷版：解压 GzipStr 的产物并还原为字符串。
func (z *utilZip) UnGzipStr(s string) (string, error) {
	out, err := z.UnGzip([]byte(s))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

/* ---------------- zlib ---------------- */

// Zlib 使用 zlib 算法压缩字节流（默认压缩级别，对应 hutool zlib 默认档位）。
// 与 Gzip 的区别：zlib 头更短、无文件级元信息，常用于协议内嵌的数据压缩。
func (z *utilZip) Zlib(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnZlib 解压 zlib 压缩的字节流（对应 hutool unZlib）。
func (z *utilZip) UnZlib(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

/* ---------------- 内部辅助 ---------------- */

// zipEntryData 待写入 zip 的条目数据：条目名（/ 分隔，目录以 / 结尾）、修改时间与内容。
type zipEntryData struct {
	name string
	mod  time.Time
	data []byte
}

// zipPack 把 srcPaths 中的文件/目录打包写入 zipPath。
// withRoot 对每个路径独立生效：为 true 时包内保留路径本身这一层目录
// （以其基本名作为包内根），为 false 时目录仅打包其下内容（对照 hutool ZipWriter.add 的 withSrcDir）。
func zipPack(zipPath string, srcPaths []string, withRoot bool) error {
	if err := zipValidateTarget(zipPath, srcPaths); err != nil {
		return err
	}
	zipAbs, err := filepath.Abs(zipPath)
	if err != nil {
		return err
	}
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for _, src := range srcPaths {
		if err := zipAddPath(w, zipAbs, src, withRoot); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return f.Close()
}

// zipValidateTarget 校验压缩目标与源路径（对照 hutool validateFiles）：
// zipPath 不能是已存在的目录；每个源路径必须存在；
// zipPath 不能位于任何作为目录的源路径内部（防止无限递归压缩）。
func zipValidateTarget(zipPath string, srcPaths []string) error {
	if info, err := os.Stat(zipPath); err == nil && info.IsDir() {
		return fmt.Errorf("zip 目标路径 [%s] 是已存在的目录", zipPath)
	}
	zipAbs, err := filepath.Abs(zipPath)
	if err != nil {
		return err
	}
	for _, src := range srcPaths {
		info, err := os.Stat(src)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			continue
		}
		srcAbs, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		if zipIsSubPath(zipAbs, srcAbs) {
			return fmt.Errorf("zip 目标路径 [%s] 不能位于被压缩目录 [%s] 内部", zipPath, src)
		}
	}
	return nil
}

// zipAddPath 把单个源文件或目录递归写入 zip 写入器 w。
// 条目名的截取起点与 hutool ZipWriter._add 的 srcRootDir 一致：
// 文件、或需要保留目录层（withRoot=true）时取其父目录，否则取目录本身（仅打包内容）。
func zipAddPath(w *zip.Writer, zipAbs, srcPath string, withRoot bool) error {
	srcPath = filepath.Clean(srcPath)
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	root := srcPath
	if !info.IsDir() || withRoot {
		root = filepath.Dir(srcPath)
	}
	return filepath.WalkDir(srcPath, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 跳过压缩结果文件本身，避免把输出压进输出（对照 hutool issue#IAGYDG）
		if abs, absErr := filepath.Abs(p); absErr == nil && abs == zipAbs {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." {
			// 源目录本身（withRoot=false 时），不产生条目
			return nil
		}
		finfo, err := d.Info()
		if err != nil {
			return err
		}
		hdr := &zip.FileHeader{
			Name:     filepath.ToSlash(rel),
			Modified: finfo.ModTime(),
		}
		hdr.SetMode(finfo.Mode())
		if finfo.IsDir() {
			// 目录条目以 / 结尾（空目录也需保留）
			hdr.Name += "/"
			_, err = w.CreateHeader(hdr)
			return err
		}
		hdr.Method = zip.Deflate
		out, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

// zipWriteAll 用 entries 整体重写 zipPath；“先读旧条目再连同新条目重写”即实现追加语义。
func zipWriteAll(zipPath string, entries []zipEntryData) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for _, e := range entries {
		if err := zipWriteEntry(w, e); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return f.Close()
}

// zipWriteEntry 写入单个数据条目。
func zipWriteEntry(w *zip.Writer, e zipEntryData) error {
	mod := e.mod
	if mod.IsZero() {
		mod = time.Now()
	}
	hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate, Modified: mod}
	if strings.HasSuffix(e.name, "/") {
		// 目录条目使用 Store 存储
		hdr.Method = zip.Store
	}
	out, err := w.CreateHeader(hdr)
	if err != nil {
		return err
	}
	if len(e.data) == 0 {
		return nil
	}
	_, err = out.Write(e.data)
	return err
}

// zipReadAllEntries 读出 zip 内全部条目（目录条目数据为 nil），用于整体重写实现追加。
func zipReadAllEntries(zipPath string) ([]zipEntryData, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	entries := make([]zipEntryData, 0, len(r.File))
	for _, f := range r.File {
		mod := f.Modified
		if mod.IsZero() {
			mod = time.Now()
		}
		if f.FileInfo().IsDir() {
			entries = append(entries, zipEntryData{name: f.Name, mod: mod})
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		entries = append(entries, zipEntryData{name: f.Name, mod: mod, data: data})
	}
	return entries, nil
}

// zipNormalizeEntryName 规范化 zip 条目名：去掉开头分隔符并统一为 / 分隔。
func zipNormalizeEntryName(name string) string {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimPrefix(name, "\\")
	return filepath.ToSlash(name)
}

// zipIsUnsafeName 判断 zip 条目名是否不安全（zip-slip）：
// 含 ".."（目录穿越）或以 "/"、"\" 开头（绝对路径）均视为攻击条目，另拒绝 Windows 盘符写法。
func zipIsUnsafeName(name string) bool {
	if name == "" {
		return true
	}
	if strings.Contains(name, "..") {
		return true
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return true
	}
	// Windows 盘符绝对路径（如 C:/evil.txt）
	return len(name) >= 2 && name[1] == ':'
}

// zipIsSubPath 判断 path 是否位于 dir 内部（含与 dir 相同）。
func zipIsSubPath(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return false
	}
	return true
}
