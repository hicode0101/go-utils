package utils

import "testing"

/* ---------------- 原有测试（自 utils_test.go 原样迁入） ---------------- */

func TestPage(t *testing.T) {
	start, end := Page.TransToStartEnd(2, 10)
	if start != 10 || end != 20 {
		t.Errorf("TransToStartEnd(2,10) = [%d,%d), want [10,20)", start, end)
	}
	if got := Page.GetStart(3, 10); got != 20 {
		t.Errorf("GetStart = %d", got)
	}
	if got := Page.TotalPage(101, 20); got != 6 {
		t.Errorf("TotalPage = %d", got)
	}
	rb := Page.Rainbow(5, 20, 7)
	if len(rb) != 7 || rb[0] != 2 || rb[6] != 8 {
		t.Errorf("Rainbow(5,20,7) = %v, want [2..8]", rb)
	}
	if rb := Page.Rainbow(1, 5, 7); len(rb) != 5 || rb[0] != 1 {
		t.Errorf("Rainbow 页数不足时 = %v", rb)
	}
}

/* ---------------- 补充覆盖 ---------------- */

func TestPageTransToStartEndNonPositive(t *testing.T) {
	// 页码 <= 0 按第 1 页处理
	start, end := Page.TransToStartEnd(0, 10)
	if start != 0 || end != 10 {
		t.Errorf("TransToStartEnd(0,10) = [%d,%d), want [0,10)", start, end)
	}
	start, end = Page.TransToStartEnd(-3, 5)
	if start != 0 || end != 5 {
		t.Errorf("TransToStartEnd(-3,5) = [%d,%d), want [0,5)", start, end)
	}
}

func TestPageGetStartEnd(t *testing.T) {
	if got := Page.GetStart(3, 10); got != 20 {
		t.Errorf("GetStart(3,10) = %d, want 20", got)
	}
	if got := Page.GetStart(0, 10); got != 0 {
		t.Errorf("GetStart(0,10) = %d, want 0（页码 <=0 按第 1 页）", got)
	}
	if got := Page.GetEnd(3, 10); got != 30 {
		t.Errorf("GetEnd(3,10) = %d, want 30", got)
	}
	if got := Page.GetEnd(1, 10); got != 10 {
		t.Errorf("GetEnd(1,10) = %d, want 10", got)
	}
}

func TestPageTotalPageDelegatesMathPageCount(t *testing.T) {
	// TotalPage 直接复用 Math.PageCount，语义应完全一致
	cases := []struct {
		total, size int64
		want        int64
	}{
		{101, 20, 6},
		{100, 20, 5},
		{0, 20, 0},
		{10, 0, 0},
	}
	for _, c := range cases {
		got, want := Page.TotalPage(c.total, c.size), Math.PageCount(c.total, c.size)
		if got != c.want || want != c.want {
			t.Errorf("TotalPage(%d,%d) = %d（Math.PageCount = %d）, want %d", c.total, c.size, got, want, c.want)
		}
	}
}

func TestPageRainbowCenter(t *testing.T) {
	rb := Page.Rainbow(5, 20, 7)
	if len(rb) != 7 || rb[0] != 2 || rb[6] != 8 {
		t.Errorf("Rainbow(5,20,7) = %v, want [2..8]（居中）", rb)
	}
	if rb := Page.Rainbow(5, 20, 1); len(rb) != 1 || rb[0] != 5 {
		t.Errorf("Rainbow(5,20,1) = %v, want [5]", rb)
	}
	if rb := Page.Rainbow(5, 20, 2); len(rb) != 2 || rb[0] != 4 || rb[1] != 5 {
		t.Errorf("Rainbow(5,20,2) = %v, want [4 5]", rb)
	}
}

func TestPageRainbowEdges(t *testing.T) {
	// 靠左：窗口自动贴边
	rb := Page.Rainbow(1, 20, 7)
	if len(rb) != 7 || rb[0] != 1 || rb[6] != 7 {
		t.Errorf("Rainbow(1,20,7) = %v, want [1..7]（贴左）", rb)
	}
	// 靠右：窗口自动贴边
	rb = Page.Rainbow(20, 20, 7)
	if len(rb) != 7 || rb[0] != 14 || rb[6] != 20 {
		t.Errorf("Rainbow(20,20,7) = %v, want [14..20]（贴右）", rb)
	}
	// 总页数不足 displayCount 时返回全部页码
	rb = Page.Rainbow(1, 5, 7)
	if len(rb) != 5 || rb[0] != 1 || rb[4] != 5 {
		t.Errorf("Rainbow(1,5,7) = %v, want [1..5]", rb)
	}
	// 页码 <= 0 按第 1 页处理
	rb = Page.Rainbow(0, 10, 3)
	if len(rb) != 3 || rb[0] != 1 || rb[2] != 3 {
		t.Errorf("Rainbow(0,10,3) = %v, want [1 2 3]", rb)
	}
}

func TestPageRainbowInvalid(t *testing.T) {
	if rb := Page.Rainbow(5, 0, 7); rb != nil {
		t.Errorf("Rainbow 总页数 <=0 应返回 nil, got %v", rb)
	}
	if rb := Page.Rainbow(5, 20, 0); rb != nil {
		t.Errorf("Rainbow displayCount <=0 应返回 nil, got %v", rb)
	}
}
