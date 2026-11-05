package utils

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// zipTestHas 断言 names 中包含 want（不依赖条目顺序）。
func zipTestHas(t *testing.T, names []string, want string) {
	t.Helper()
	for _, n := range names {
		if n == want {
			return
		}
	}
	t.Errorf("条目列表 %v 中应包含 %q", names, want)
}

// zipTestMustContain 读取文件并断言内容一致。
func zipTestMustContain(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("文件 %s 内容 = %q, want %q", path, got, want)
	}
}

// zipTestTree 构造带嵌套子目录的源目录：
//
//	src/a.txt
//	src/nested/b.txt
//	src/nested/deep/c.txt
//	src/emptydir/（空目录）
func zipTestTree(t *testing.T) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "src")
	for dir, file := range map[string]string{
		"":                              "a.txt",
		"nested":                        "b.txt",
		filepath.Join("nested", "deep"): "c.txt",
	} {
		if err := os.MkdirAll(filepath.Join(src, dir), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, dir, file), []byte("content of "+file), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(src, "emptydir"), 0755); err != nil {
		t.Fatal(err)
	}
	return src
}

// TestZipDirWithRoot ZipDir withRoot=true：包内保留源目录这一层，嵌套结构完整。
func TestZipDirWithRoot(t *testing.T) {
	src := zipTestTree(t)
	zipPath := filepath.Join(t.TempDir(), "withroot.zip")
	if err := Zip.ZipDir(src, zipPath, true); err != nil {
		t.Fatalf("ZipDir 失败: %v", err)
	}
	// 条目名以源目录基本名 src 开头
	if got, err := Zip.ListFileNames(zipPath, "src"); err != nil || len(got) != 1 {
		t.Errorf("ListFileNames(src) = %v, %v; want [a.txt]", got, err)
	} else {
		zipTestHas(t, got, "a.txt")
	}
	if got, err := Zip.ListFileNames(zipPath, "src/nested"); err != nil || len(got) != 1 {
		t.Errorf("ListFileNames(src/nested) = %v, %v; want [b.txt]", got, err)
	} else {
		zipTestHas(t, got, "b.txt")
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "src/nested/deep/c.txt"); err != nil || string(entry) != "content of c.txt" {
		t.Errorf("ReadEntryBytes = %q, %v", entry, err)
	}
}

// TestZipDirWithoutRoot ZipDir withRoot=false：包内仅包含目录下的内容，不含源目录层。
func TestZipDirWithoutRoot(t *testing.T) {
	src := zipTestTree(t)
	zipPath := filepath.Join(t.TempDir(), "noroot.zip")
	if err := Zip.ZipDir(src, zipPath, false); err != nil {
		t.Fatalf("ZipDir 失败: %v", err)
	}
	// 根目录直接列出顶层文件，目录条目与子目录内容均不出现
	if got, err := Zip.ListFileNames(zipPath, ""); err != nil || len(got) != 1 {
		t.Errorf("ListFileNames(\"\") = %v, %v; want [a.txt]", got, err)
	} else {
		zipTestHas(t, got, "a.txt")
	}
	if got, err := Zip.ListFileNames(zipPath, "nested"); err != nil || len(got) != 1 {
		t.Errorf("ListFileNames(nested) = %v, %v; want [b.txt]", got, err)
	} else {
		zipTestHas(t, got, "b.txt")
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "nested/deep/c.txt"); err != nil || string(entry) != "content of c.txt" {
		t.Errorf("ReadEntryBytes = %q, %v", entry, err)
	}
}

// TestZipDirErrors ZipDir 非法输入：源不存在、源是文件、输出目标非法。
func TestZipDirErrors(t *testing.T) {
	dir := t.TempDir()
	// 源目录不存在
	if err := Zip.ZipDir(filepath.Join(dir, "nope"), filepath.Join(dir, "a.zip"), false); err == nil {
		t.Error("源目录不存在应返回错误")
	}
	// 源路径是文件而非目录
	file := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Zip.ZipDir(file, filepath.Join(dir, "b.zip"), false); err == nil {
		t.Error("源路径是文件应返回错误")
	}
	// zip 目标位于被压缩目录内部（validateFiles 防无限递归）
	src := zipTestTree(t)
	if err := Zip.ZipDir(src, filepath.Join(src, "out.zip"), false); err == nil {
		t.Error("zip 目标在被压缩目录内部应返回错误")
	}
	// zip 目标是已存在的目录
	if err := Zip.ZipDir(src, dir, false); err == nil {
		t.Error("zip 目标是已存在的目录应返回错误")
	}
}

// TestZipFilesMulti ZipFiles：多文件与目录混合打包，各自以基本名出现在包根目录。
func TestZipFilesMulti(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "log1.txt")
	f2 := filepath.Join(dir, "log2.txt")
	if err := os.WriteFile(f1, []byte("one"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("two"), 0644); err != nil {
		t.Fatal(err)
	}
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(filepath.Join(assets, "img"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "img", "logo.png"), []byte("png-bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(dir, "multi.zip")
	if err := Zip.ZipFiles(zipPath, f1, f2, assets); err != nil {
		t.Fatalf("ZipFiles 失败: %v", err)
	}
	// 包根目录下的文件（目录条目不列入）
	root, err := Zip.ListFileNames(zipPath, "")
	if err != nil {
		t.Fatalf("ListFileNames 失败: %v", err)
	}
	zipTestHas(t, root, "log1.txt")
	zipTestHas(t, root, "log2.txt")
	if len(root) != 2 {
		t.Errorf("根目录文件数 = %d, want 2: %v", len(root), root)
	}
	// 目录连同内部结构一起打包（保留目录本身这一层）
	if entry, err := Zip.ReadEntryBytes(zipPath, "log1.txt"); err != nil || string(entry) != "one" {
		t.Errorf("ReadEntryBytes(log1.txt) = %q, %v", entry, err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "assets/img/logo.png"); err != nil || string(entry) != "png-bytes" {
		t.Errorf("ReadEntryBytes(assets/img/logo.png) = %q, %v", entry, err)
	}
}

// TestZipFilesErrors ZipFiles 非法输入：空列表、目标在被压缩目录内、目标是目录、源不存在。
func TestZipFilesErrors(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(dir, "out.zip")
	// 未提供任何待压缩路径
	if err := Zip.ZipFiles(zipPath); err == nil {
		t.Error("空路径列表应返回错误")
	}
	// zip 目标位于被压缩目录内部（对照 hutool validateFiles）
	if err := Zip.ZipFiles(filepath.Join(dir, "inner.zip"), dir); err == nil {
		t.Error("zip 目标在被压缩目录内部应返回错误")
	}
	// zip 目标是已存在的目录
	targetDir := filepath.Join(dir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Zip.ZipFiles(targetDir, f); err == nil {
		t.Error("zip 目标是已存在的目录应返回错误")
	}
	// 源路径不存在
	if err := Zip.ZipFiles(zipPath, filepath.Join(dir, "missing.txt")); err == nil {
		t.Error("源文件不存在应返回错误")
	}
}

// TestZipEntryBytes ZipEntryBytes：新建压缩包、追加条目、同名保留旧值、反斜杠条目名规范化。
func TestZipEntryBytes(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "entry.zip")
	// zip 不存在：新建仅含该条目的压缩包
	if err := Zip.ZipEntryBytes(zipPath, "docs/readme.txt", []byte("v1")); err != nil {
		t.Fatalf("ZipEntryBytes 新建失败: %v", err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "docs/readme.txt"); err != nil || string(entry) != "v1" {
		t.Errorf("ReadEntryBytes = %q, %v", entry, err)
	}
	// zip 已存在：追加新条目，旧条目保留
	if err := Zip.ZipEntryBytes(zipPath, "docs/other.txt", []byte("v2")); err != nil {
		t.Fatalf("ZipEntryBytes 追加失败: %v", err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "docs/other.txt"); err != nil || string(entry) != "v2" {
		t.Errorf("追加后 ReadEntryBytes = %q, %v", entry, err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "docs/readme.txt"); err != nil || string(entry) != "v1" {
		t.Errorf("追加后旧条目应保留, got %q, %v", entry, err)
	}
	// 包内已存在同名条目：保留原条目、跳过写入（对照 hutool append 默认不覆盖）
	if err := Zip.ZipEntryBytes(zipPath, "docs/readme.txt", []byte("v9")); err != nil {
		t.Fatalf("ZipEntryBytes 同名写入失败: %v", err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "docs/readme.txt"); err != nil || string(entry) != "v1" {
		t.Errorf("同名条目应保留旧值 v1, got %q, %v", entry, err)
	}
	// 条目名反斜杠自动转换为 /
	if entry, err := Zip.ReadEntryBytes(zipPath, "docs\\readme.txt"); err != nil || string(entry) != "v1" {
		t.Errorf("反斜杠条目名应命中: %q, %v", entry, err)
	}
	// 空条目路径（"/" 规范化后为空）
	if err := Zip.ZipEntryBytes(zipPath, "/", []byte("x")); err == nil {
		t.Error("空条目路径应返回错误")
	}
}

// TestZipUnzipRoundTrip Unzip：目录打包解压往返，嵌套目录、中文文件名与空目录均还原。
func TestZipUnzipRoundTrip(t *testing.T) {
	src := zipTestTree(t)
	zhFile := filepath.Join(src, "中文文件.txt")
	if err := os.WriteFile(zhFile, []byte("中文内容"), 0644); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(t.TempDir(), "rt.zip")
	if err := Zip.ZipDir(src, zipPath, false); err != nil {
		t.Fatalf("ZipDir 失败: %v", err)
	}
	outDir := filepath.Join(t.TempDir(), "out")
	if err := Zip.Unzip(zipPath, outDir); err != nil {
		t.Fatalf("Unzip 失败: %v", err)
	}
	zipTestMustContain(t, filepath.Join(outDir, "a.txt"), "content of a.txt")
	zipTestMustContain(t, filepath.Join(outDir, "nested", "b.txt"), "content of b.txt")
	zipTestMustContain(t, filepath.Join(outDir, "nested", "deep", "c.txt"), "content of c.txt")
	zipTestMustContain(t, filepath.Join(outDir, "中文文件.txt"), "中文内容")
	// 空目录还原
	if info, err := os.Stat(filepath.Join(outDir, "emptydir")); err != nil || !info.IsDir() {
		t.Errorf("空目录应被还原: %v, %v", info, err)
	}
}

// TestZipUnzipZipSlip Unzip zip-slip 防护：条目名含 .. 或绝对路径时拒绝解压且不落盘。
func TestZipUnzipZipSlip(t *testing.T) {
	for _, name := range []string{
		"../evil.txt",        // 目录穿越
		"sub/../../evil.txt", // 中途穿越
		"/abs.txt",           // Unix 绝对路径
		"\\abs.txt",          // 反斜杠绝对路径
		"C:/evil.txt",        // Windows 盘符
	} {
		var buf bytes.Buffer
		w := zip.NewWriter(&buf)
		hdr := &zip.FileHeader{Name: name, Method: zip.Store}
		fw, err := w.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte("evil")); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		zipPath := filepath.Join(t.TempDir(), "slip.zip")
		if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		outDir := filepath.Join(t.TempDir(), "out")
		if err := Zip.Unzip(zipPath, outDir); err == nil {
			t.Errorf("条目 %q 应触发 zip-slip 防护返回错误", name)
		}
		// 解压目录内不应有任何落盘文件
		if entries, err := os.ReadDir(outDir); err == nil && len(entries) > 0 {
			t.Errorf("条目 %q 被拒绝后解压目录应为空, got %v", name, entries)
		}
		// ../evil.txt 不应出现在解压目录的父目录
		if _, err := os.Stat(filepath.Join(outDir, "..", "evil.txt")); err == nil {
			t.Errorf("条目 %q 竟然落盘到了解压目录之外", name)
		}
	}
}

// TestZipUnzipCorrupted Unzip/ReadEntryBytes 对损坏 zip 返回 error：
// 伪造内容与被截断的合法 zip 均应拒绝（损坏 zip 的读取行为与加密 zip 一致地报错）。
func TestZipUnzipCorrupted(t *testing.T) {
	dir := t.TempDir()
	// 伪造内容
	bad := filepath.Join(dir, "bad.zip")
	if err := os.WriteFile(bad, []byte("this is definitely not a zip file"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Zip.Unzip(bad, filepath.Join(dir, "out1")); err == nil {
		t.Error("非 zip 内容应返回错误")
	}
	if _, err := Zip.ReadEntryBytes(bad, "x.txt"); err == nil {
		t.Error("非 zip 内容 ReadEntryBytes 应返回错误")
	}
	// 合法 zip 截断一半（中央目录位于文件尾部，截断后即损坏）
	good := filepath.Join(dir, "good.zip")
	if err := Zip.ZipEntryBytes(good, "a.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(good)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(good, info.Size()/2); err != nil {
		t.Fatal(err)
	}
	if err := Zip.Unzip(good, filepath.Join(dir, "out2")); err == nil {
		t.Error("截断的 zip 应返回错误")
	}
	if _, err := Zip.ReadEntryBytes(good, "a.txt"); err == nil {
		t.Error("截断的 zip ReadEntryBytes 应返回错误")
	}
	// zip 文件本身不存在
	if err := Zip.Unzip(filepath.Join(dir, "missing.zip"), filepath.Join(dir, "out3")); err == nil {
		t.Error("zip 文件不存在应返回错误")
	}
}

// TestZipReadEntryBytes ReadEntryBytes：命中、反斜杠规范化、目录条目、不存在与不存在的 zip 文件。
func TestZipReadEntryBytes(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "read.zip")
	if err := Zip.ZipEntryBytes(zipPath, "logs/app.log", []byte("log line")); err != nil {
		t.Fatal(err)
	}
	if err := Zip.ZipEntryBytes(zipPath, "logs/", nil); err != nil { // 目录条目
		t.Fatal(err)
	}
	if entry, err := Zip.ReadEntryBytes(zipPath, "logs/app.log"); err != nil || string(entry) != "log line" {
		t.Errorf("命中条目 = %q, %v", entry, err)
	}
	// Windows 反斜杠路径自动转换
	if entry, err := Zip.ReadEntryBytes(zipPath, "logs\\app.log"); err != nil || string(entry) != "log line" {
		t.Errorf("反斜杠条目名应命中: %q, %v", entry, err)
	}
	// 目录条目无内容可读
	if _, err := Zip.ReadEntryBytes(zipPath, "logs/"); err == nil {
		t.Error("目录条目应返回错误")
	}
	// 条目不存在
	if _, err := Zip.ReadEntryBytes(zipPath, "logs/nope.log"); err == nil {
		t.Error("不存在的条目应返回错误")
	}
	// 空条目路径
	if _, err := Zip.ReadEntryBytes(zipPath, ""); err == nil {
		t.Error("空条目路径应返回错误")
	}
	// zip 文件不存在
	if _, err := Zip.ReadEntryBytes(filepath.Join(t.TempDir(), "missing.zip"), "a.txt"); err == nil {
		t.Error("zip 文件不存在应返回错误")
	}
}

// TestZipListFileNames ListFileNames：dir 前缀过滤、尾斜杠可省略、空 dir 全列、未知 dir 返回空。
func TestZipListFileNames(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "list.zip")
	entries := []struct {
		name string
		data string
	}{
		{"a.txt", "a"},
		{"b.txt", "b"},
		{"sub/", ""},
		{"sub/in.txt", "in"},
		{"sub/deep/", ""},
		{"sub/deep/c.txt", "c"},
	}
	for _, e := range entries {
		if err := Zip.ZipEntryBytes(zipPath, e.name, []byte(e.data)); err != nil {
			t.Fatalf("ZipEntryBytes(%s) 失败: %v", e.name, err)
		}
	}
	// 空 dir：根目录下全部文件（目录条目与子目录内容不出现）
	got, err := Zip.ListFileNames(zipPath, "")
	if err != nil {
		t.Fatalf("ListFileNames 失败: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("根目录文件 = %v, want [a.txt b.txt]", got)
	}
	zipTestHas(t, got, "a.txt")
	zipTestHas(t, got, "b.txt")
	// dir 前缀过滤：仅当前层级的文件
	if got, err = Zip.ListFileNames(zipPath, "sub"); err != nil || len(got) != 1 || got[0] != "in.txt" {
		t.Errorf("ListFileNames(sub) = %v, %v; want [in.txt]", got, err)
	}
	// 尾部 "/" 可省略，两种写法等价
	if got2, err := Zip.ListFileNames(zipPath, "sub/"); err != nil || len(got2) != len(got) {
		t.Errorf("ListFileNames(sub/) = %v, %v; want 与 sub 一致", got2, err)
	}
	if got, err = Zip.ListFileNames(zipPath, "sub/deep"); err != nil || len(got) != 1 || got[0] != "c.txt" {
		t.Errorf("ListFileNames(sub/deep) = %v, %v; want [c.txt]", got, err)
	}
	// 不存在的 dir：空切片
	if got, err = Zip.ListFileNames(zipPath, "no_such_dir"); err != nil || len(got) != 0 {
		t.Errorf("不存在的 dir 应返回空切片, got %v, %v", got, err)
	}
	// zip 文件不存在
	if _, err = Zip.ListFileNames(filepath.Join(t.TempDir(), "missing.zip"), ""); err == nil {
		t.Error("zip 文件不存在应返回错误")
	}
}

// TestZipGzipRoundTrip Gzip/UnGzip 往返：含中文内容与空切片。
func TestZipGzipRoundTrip(t *testing.T) {
	data := []byte("gzip compress test 中文测试")
	gz, err := Zip.Gzip(data)
	if err != nil {
		t.Fatalf("Gzip 失败: %v", err)
	}
	if len(gz) == 0 || bytes.Equal(gz, data) {
		t.Error("Gzip 结果应为非空且不同于原文")
	}
	back, err := Zip.UnGzip(gz)
	if err != nil || string(back) != string(data) {
		t.Errorf("gzip 往返失败: %q, %v", back, err)
	}
	// 空切片
	gzEmpty, err := Zip.Gzip(nil)
	if err != nil {
		t.Fatalf("Gzip(nil) 失败: %v", err)
	}
	backEmpty, err := Zip.UnGzip(gzEmpty)
	if err != nil || len(backEmpty) != 0 {
		t.Errorf("空切片往返 = %v, %v", backEmpty, err)
	}
	// 非法输入
	if _, err := Zip.UnGzip([]byte("not gzip data")); err == nil {
		t.Error("非 gzip 数据应返回错误")
	}
	if _, err := Zip.UnGzip(nil); err == nil {
		t.Error("空输入应返回错误")
	}
}

// TestZipGzipStrRoundTrip GzipStr/UnGzipStr 字符串往返。
func TestZipGzipStrRoundTrip(t *testing.T) {
	const s = "需要压缩的很长很长的字符串 with ascii"
	gz, err := Zip.GzipStr(s)
	if err != nil {
		t.Fatalf("GzipStr 失败: %v", err)
	}
	if gz == "" || gz == s {
		t.Error("GzipStr 结果应为非空且不同于原文")
	}
	back, err := Zip.UnGzipStr(gz)
	if err != nil || back != s {
		t.Errorf("GzipStr/UnGzipStr 往返失败: %q, %v", back, err)
	}
	// 字符串版非法输入
	if _, err := Zip.UnGzipStr("plain text"); err == nil {
		t.Error("非 gzip 字符串应返回错误")
	}
}

// TestZipZlibRoundTrip Zlib/UnZlib 往返：含中文内容与空切片，及非法输入。
func TestZipZlibRoundTrip(t *testing.T) {
	data := []byte("zlib compress test 中文测试")
	zl, err := Zip.Zlib(data)
	if err != nil {
		t.Fatalf("Zlib 失败: %v", err)
	}
	if len(zl) == 0 || bytes.Equal(zl, data) {
		t.Error("Zlib 结果应为非空且不同于原文")
	}
	back, err := Zip.UnZlib(zl)
	if err != nil || string(back) != string(data) {
		t.Errorf("zlib 往返失败: %q, %v", back, err)
	}
	// 空切片
	zlEmpty, err := Zip.Zlib(nil)
	if err != nil {
		t.Fatalf("Zlib(nil) 失败: %v", err)
	}
	backEmpty, err := Zip.UnZlib(zlEmpty)
	if err != nil || len(backEmpty) != 0 {
		t.Errorf("空切片往返 = %v, %v", backEmpty, err)
	}
	// 非法输入
	if _, err := Zip.UnZlib([]byte("not zlib data")); err == nil {
		t.Error("非 zlib 数据应返回错误")
	}
	if _, err := Zip.UnZlib(nil); err == nil {
		t.Error("空输入应返回错误")
	}
}
