package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

func TestFile(t *testing.T) {
	dir := t.TempDir()
	path := File.Join(dir, "a.txt")

	if err := File.WriteText(path, "hello"); err != nil {
		t.Fatalf("WriteText 失败: %v", err)
	}
	if got, _ := File.ReadText(path); got != "hello" {
		t.Errorf("ReadText = %q", got)
	}
	if err := File.AppendText(path, "world"); err != nil {
		t.Fatalf("AppendText 失败: %v", err)
	}
	if got, _ := File.ReadText(path); got != "helloworld\n" {
		t.Errorf("Append 后内容 = %q", got)
	}
	if got := File.Size(path); got != int64(len("helloworld\n")) {
		t.Errorf("Size = %d", got)
	}
	if !File.Exists(path) {
		t.Error("文件应存在")
	}

	dst := File.Join(dir, "b.txt")
	if err := File.Copy(path, dst); err != nil {
		t.Fatalf("Copy 失败: %v", err)
	}
	if !File.Exists(dst) {
		t.Error("复制后的文件应存在")
	}

	files, err := File.ListFiles(dir, ".txt")
	if err != nil || len(files) != 2 {
		t.Errorf("ListFiles = %v, %v; want 2 个 txt 文件", files, err)
	}

	if got := File.Name("/app/src/main.go"); got != "main.go" {
		t.Errorf("Name = %q", got)
	}
	if got := File.NameWithoutExt("/app/src/main.go"); got != "main" {
		t.Errorf("NameWithoutExt = %q", got)
	}
	if got := File.Dir("/app/src/main.go"); got != File.Join("/app", "src") {
		t.Errorf("Dir = %q", got)
	}
}

/* ---------------- Create / Open ---------------- */

func TestFileCreateOpen(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "create.txt")
	// Create 新建并可写入
	f, err := File.Create(p)
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if _, err := f.WriteString("abc"); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	f.Close()
	if got := File.Size(p); got != 3 {
		t.Errorf("Create 后 Size = %d, want 3", got)
	}
	// Create 已存在则截断清空
	f2, err := File.Create(p)
	if err != nil {
		t.Fatalf("Create 已存在文件失败: %v", err)
	}
	f2.Close()
	if got := File.Size(p); got != 0 {
		t.Errorf("Create 应截断为空, Size = %d", got)
	}
	// Open 只读打开
	f3, err := File.Open(p)
	if err != nil {
		t.Fatalf("Open 失败: %v", err)
	}
	f3.Close()
	// Open 不存在的文件应报错
	if _, err := File.Open(File.Join(dir, "nope.txt")); err == nil {
		t.Error("Open 不存在文件应返回错误")
	}
}

/* ---------------- OpenFile（追加模式验证） ---------------- */

func TestFileOpenFileAppend(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "flag.txt")
	if err := File.WriteText(p, "a"); err != nil {
		t.Fatal(err)
	}
	f, err := File.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile 失败: %v", err)
	}
	if _, err := f.WriteString("b"); err != nil {
		t.Fatalf("追加写入失败: %v", err)
	}
	f.Close()
	if got, _ := File.ReadText(p); got != "ab" {
		t.Errorf("O_APPEND 应追加而非覆盖, got %q", got)
	}
	// 无 O_CREATE 打开不存在文件应报错
	if _, err := File.OpenFile(File.Join(dir, "nope.txt"), os.O_WRONLY, 0644); err == nil {
		t.Error("无 O_CREATE 打开不存在文件应返回错误")
	}
}

/* ---------------- Read / ReadText / Write / WriteText ---------------- */

func TestFileReadWrite(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "rw.bin")
	if err := File.Write(p, []byte{0x00, 0x01, 0xFF}); err != nil {
		t.Fatalf("Write 失败: %v", err)
	}
	data, err := File.Read(p)
	if err != nil || len(data) != 3 || data[0] != 0x00 || data[2] != 0xFF {
		t.Errorf("Read = %v, %v", data, err)
	}
	// Write 覆盖旧内容
	if err := File.WriteText(p, "hello 文本"); err != nil {
		t.Fatalf("WriteText 失败: %v", err)
	}
	if got, err := File.ReadText(p); err != nil || got != "hello 文本" {
		t.Errorf("ReadText = %q, %v", got, err)
	}
	if err := File.WriteText(p, "x"); err != nil {
		t.Fatal(err)
	}
	if got, _ := File.ReadText(p); got != "x" {
		t.Errorf("WriteText 应覆盖旧内容, got %q", got)
	}
	// 读不存在文件
	if _, err := File.Read(File.Join(dir, "nope.bin")); err == nil {
		t.Error("Read 不存在文件应返回错误")
	}
	if got, err := File.ReadText(File.Join(dir, "nope.txt")); err == nil || got != "" {
		t.Error("ReadText 不存在文件应返回错误与空串")
	}
}

/* ---------------- Append / AppendText ---------------- */

func TestFileAppendText(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "log.txt")
	// Append 原始字节（无换行）
	if err := File.Append(p, []byte("line1")); err != nil {
		t.Fatalf("Append 失败: %v", err)
	}
	// AppendText 自动补换行
	if err := File.AppendText(p, "line2"); err != nil {
		t.Fatalf("AppendText 失败: %v", err)
	}
	if got, _ := File.ReadText(p); got != "line1line2\n" {
		t.Errorf("追加结果 = %q, want \"line1line2\\n\"", got)
	}
	// 已带换行不重复补
	if err := File.AppendText(p, "line3\n"); err != nil {
		t.Fatal(err)
	}
	if got, _ := File.ReadText(p); got != "line1line2\nline3\n" {
		t.Errorf("带换行追加 = %q, want \"line1line2\\nline3\\n\"", got)
	}
	// Append 到不存在的文件会创建
	p2 := File.Join(dir, "new.txt")
	if err := File.Append(p2, []byte("x")); err != nil {
		t.Fatalf("Append 新文件失败: %v", err)
	}
	if !File.Exists(p2) {
		t.Error("Append 应创建不存在的文件")
	}
}

/* ---------------- Copy ---------------- */

func TestFileCopy(t *testing.T) {
	dir := t.TempDir()
	src := File.Join(dir, "src.txt")
	if err := File.WriteText(src, "copy me 你好"); err != nil {
		t.Fatal(err)
	}
	dst := File.Join(dir, "dst.txt")
	if err := File.Copy(src, dst); err != nil {
		t.Fatalf("Copy 失败: %v", err)
	}
	if got, _ := File.ReadText(dst); got != "copy me 你好" {
		t.Errorf("复制内容 = %q", got)
	}
	// 权限保留
	srcInfo, err := File.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	dstInfo, err := File.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if dstInfo.Mode().Perm() != srcInfo.Mode().Perm() {
		t.Errorf("复制后权限应保留: src=%v dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
	// 覆盖已存在的目标
	if err := File.Copy(src, dst); err != nil {
		t.Errorf("覆盖复制不应报错: %v", err)
	}
	// 源不存在报错
	if err := File.Copy(File.Join(dir, "nope"), dst); err == nil {
		t.Error("复制不存在的源文件应返回错误")
	}
}

/* ---------------- Exists / Stat / Size ---------------- */

func TestFileExistsStatSize(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "info.txt")
	if File.Exists(p) {
		t.Error("不存在的文件 Exists 应为 false")
	}
	if err := File.WriteText(p, "12345"); err != nil {
		t.Fatal(err)
	}
	if !File.Exists(p) {
		t.Error("存在的文件 Exists 应为 true")
	}
	if !File.Exists(dir) {
		t.Error("存在的目录 Exists 应为 true")
	}
	if File.Exists(File.Join(dir, "nope")) {
		t.Error("不存在的路径 Exists 应为 false")
	}
	info, err := File.Stat(p)
	if err != nil {
		t.Fatalf("Stat 失败: %v", err)
	}
	if info.Size() != 5 {
		t.Errorf("Stat.Size = %d, want 5", info.Size())
	}
	if info.ModTime().IsZero() {
		t.Error("Stat.ModTime 不应为零值")
	}
	if _, err := File.Stat(File.Join(dir, "nope")); err == nil {
		t.Error("Stat 不存在文件应返回错误")
	}
	if got := File.Size(p); got != 5 {
		t.Errorf("Size = %d, want 5", got)
	}
	if got := File.Size(dir); got != -1 {
		t.Errorf("目录 Size 应为 -1, got %d", got)
	}
	if got := File.Size(File.Join(dir, "nope")); got != -1 {
		t.Errorf("不存在文件 Size 应为 -1, got %d", got)
	}
}

/* ---------------- MkdirAll ---------------- */

func TestFileMkdirAll(t *testing.T) {
	dir := t.TempDir()
	nested := File.Join(dir, "a", "b", "c")
	if err := File.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll 嵌套创建失败: %v", err)
	}
	if !File.Exists(nested) {
		t.Error("嵌套目录应存在")
	}
	// 已存在时幂等
	if err := File.MkdirAll(nested, 0755); err != nil {
		t.Errorf("MkdirAll 已存在目录应幂等: %v", err)
	}
}

/* ---------------- ListFiles ---------------- */

func TestFileListFiles(t *testing.T) {
	dir := t.TempDir()
	sub := File.Join(dir, "sub")
	if err := File.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.txt", "B.TXT", "c.go"} {
		if err := File.WriteText(File.Join(dir, name), "x"); err != nil {
			t.Fatal(err)
		}
	}
	if err := File.WriteText(File.Join(sub, "deep.txt"), "x"); err != nil {
		t.Fatal(err)
	}
	// 按扩展名过滤（大小写不敏感）并递归子目录
	files, err := File.ListFiles(dir, ".txt")
	if err != nil {
		t.Fatalf("ListFiles 失败: %v", err)
	}
	if len(files) != 3 { // a.txt、B.TXT、sub/deep.txt
		t.Errorf("过滤 .txt 应命中 3 个（含大写 .TXT 与子目录）, got %v", files)
	}
	// 空扩展名返回全部文件（递归）
	all, err := File.ListFiles(dir, "")
	if err != nil || len(all) != 4 {
		t.Errorf("无过滤应返回 4 个文件, got %v, %v", all, err)
	}
	// 空目录返回 0 个文件
	empty := File.Join(dir, "empty")
	if err := File.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}
	if got, err := File.ListFiles(empty, ""); err != nil || len(got) != 0 {
		t.Errorf("空目录应返回 0 个文件, got %v, %v", got, err)
	}
	// 不存在的目录报错
	if _, err := File.ListFiles(File.Join(dir, "nope"), ""); err == nil {
		t.Error("不存在的目录应返回错误")
	}
}

/* ---------------- Rename / Remove / RemoveAll ---------------- */

func TestFileRenameRemove(t *testing.T) {
	dir := t.TempDir()
	p := File.Join(dir, "old.txt")
	if err := File.WriteText(p, "data"); err != nil {
		t.Fatal(err)
	}
	np := File.Join(dir, "new.txt")
	if err := File.Rename(p, np); err != nil {
		t.Fatalf("Rename 失败: %v", err)
	}
	if File.Exists(p) {
		t.Error("重命名后原文件不应存在")
	}
	if !File.Exists(np) {
		t.Error("重命名后新文件应存在")
	}
	// Remove
	if err := File.Remove(np); err != nil {
		t.Fatalf("Remove 失败: %v", err)
	}
	if File.Exists(np) {
		t.Error("删除后文件不应存在")
	}
	if err := File.Remove(np); err == nil {
		t.Error("删除不存在文件应返回错误")
	}
	// RemoveAll 递归删除
	tree := File.Join(dir, "tree", "inner")
	if err := File.MkdirAll(tree, 0755); err != nil {
		t.Fatal(err)
	}
	if err := File.WriteText(File.Join(tree, "f.txt"), "x"); err != nil {
		t.Fatal(err)
	}
	if err := File.RemoveAll(File.Join(dir, "tree")); err != nil {
		t.Fatalf("RemoveAll 失败: %v", err)
	}
	if File.Exists(File.Join(dir, "tree")) {
		t.Error("RemoveAll 后目录不应存在")
	}
	if err := File.RemoveAll(File.Join(dir, "nope_dir")); err != nil {
		t.Errorf("RemoveAll 不存在路径应返回 nil: %v", err)
	}
}

/* ---------------- 路径处理（Windows 与 Unix 风格 filepath 语义） ---------------- */

func TestFilePathOps(t *testing.T) {
	// Join：以 filepath 语义为基准
	if got := File.Join("logs", "2024", "app.log"); got != filepath.Join("logs", "2024", "app.log") {
		t.Errorf("Join = %q, want %q", got, filepath.Join("logs", "2024", "app.log"))
	}
	if got := File.Join("a"); got != "a" {
		t.Errorf("Join 单元素 = %q", got)
	}
	if got := File.Join(); got != "" {
		t.Errorf("Join 无元素 = %q, want \"\"", got)
	}
	// Windows 反斜杠风格路径（仅 Windows 具备反斜杠分隔符语义）
	if runtime.GOOS == "windows" {
		if got := File.Dir(`C:\app\src\main.go`); got != `C:\app\src` {
			t.Errorf("Dir(Windows 路径) = %q, want \"C:\\\\app\\\\src\"", got)
		}
		if got := File.Join(`C:\app`, "src", "main.go"); got != `C:\app\src\main.go` {
			t.Errorf("Join(Windows 路径) = %q", got)
		}
	} else {
		// Unix 正斜杠风格路径
		if got := File.Dir("/app/src/main.go"); got != "/app/src" {
			t.Errorf("Dir(Unix 路径) = %q, want \"/app/src\"", got)
		}
		if got := File.Join("/app", "src", "main.go"); got != "/app/src/main.go" {
			t.Errorf("Join(Unix 路径) = %q", got)
		}
	}
	// Name / NameWithoutExt / Ext：正斜杠在两个平台均为分隔符，断言一致
	if got := File.Name("/app/src/main.go"); got != "main.go" {
		t.Errorf("Name = %q, want \"main.go\"", got)
	}
	if got := File.NameWithoutExt("/app/src/main.go"); got != "main" {
		t.Errorf("NameWithoutExt = %q, want \"main\"", got)
	}
	if got := File.Ext("/app/src/main.go"); got != ".go" {
		t.Errorf("Ext = %q, want \".go\"", got)
	}
	if got := File.Ext("/app/archive.tar.gz"); got != ".gz" {
		t.Errorf("Ext 多重扩展名 = %q, want \".gz\"", got)
	}
	if got := File.NameWithoutExt("readme"); got != "readme" {
		t.Errorf("NameWithoutExt 无扩展名 = %q, want \"readme\"", got)
	}
	if got := File.Ext("readme"); got != "" {
		t.Errorf("Ext 无扩展名 = %q, want \"\"", got)
	}
}

/* ---------------- 程序路径 ---------------- */

func TestFileProgramPaths(t *testing.T) {
	wd := File.WorkDir()
	if wd == "" {
		t.Error("WorkDir 不应为空")
	}
	cur := File.GetCurrentDir()
	if cur == "" {
		t.Error("GetCurrentDir 不应为空")
	}
	if wd != cur {
		t.Errorf("WorkDir(%q) 应与 GetCurrentDir(%q) 相等", wd, cur)
	}
	exe := File.ExePath()
	if exe == "" || !filepath.IsAbs(exe) {
		t.Errorf("ExePath 应为非空绝对路径, got %q", exe)
	}
	curExe := File.GetCurrentExe()
	if curExe == "" {
		t.Error("GetCurrentExe 不应为空")
	}
	if exe != curExe {
		t.Errorf("ExePath(%q) 应与 GetCurrentExe(%q) 相等", exe, curExe)
	}
}
