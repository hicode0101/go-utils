package utils

import "testing"

/* ---------------- 原有测试（自 utils_test.go 原样迁入） ---------------- */

func TestMath(t *testing.T) {
	if got := Math.PageCount(101, 20); got != 6 {
		t.Errorf("PageCount(101,20) = %d, want 6", got)
	}
	if got := Math.PageCount(0, 20); got != 0 {
		t.Errorf("PageCount(0,20) = %d, want 0", got)
	}
	if got := Math.PageCount(10, 0); got != 0 {
		t.Errorf("PageCount(10,0) = %d, want 0（除零保护）", got)
	}
	if got := Math.RoundToInt(5.5); got != 6 {
		t.Errorf("RoundToInt(5.5) = %d, want 6", got)
	}
	if got := Math.CeilToInt(5.3); got != 6 {
		t.Errorf("CeilToInt(5.3) = %d, want 6", got)
	}
	if got := Math.RoundFloat(3.14159, 2); got != 3.14 {
		t.Errorf("RoundFloat(3.14159,2) = %v, want 3.14", got)
	}
	if got := Math.AbsInt(-7); got != 7 {
		t.Errorf("AbsInt(-7) = %d, want 7", got)
	}
}

/* ---------------- 补充覆盖 ---------------- */

func TestMathAbs(t *testing.T) {
	if got := Math.Abs(-3.5); got != 3.5 {
		t.Errorf("Abs(-3.5) = %v, want 3.5", got)
	}
	if got := Math.Abs(3.5); got != 3.5 {
		t.Errorf("Abs(3.5) = %v, want 3.5", got)
	}
}

func TestMathAbsIntInt64(t *testing.T) {
	if got := Math.AbsInt(7); got != 7 {
		t.Errorf("AbsInt(7) = %d, want 7（正数原样返回）", got)
	}
	if got := Math.AbsInt(-7); got != 7 {
		t.Errorf("AbsInt(-7) = %d, want 7", got)
	}
	if got := Math.AbsInt64(-99); got != 99 {
		t.Errorf("AbsInt64(-99) = %d, want 99", got)
	}
	if got := Math.AbsInt64(99); got != 99 {
		t.Errorf("AbsInt64(99) = %d, want 99", got)
	}
}

func TestMathCeilFloorToInt(t *testing.T) {
	if got := Math.CeilToInt(5.3); got != 6 {
		t.Errorf("CeilToInt(5.3) = %d, want 6", got)
	}
	if got := Math.CeilToInt(-5.3); got != -5 {
		t.Errorf("CeilToInt(-5.3) = %d, want -5（负数朝正无穷取整）", got)
	}
	if got := Math.FloorToInt(5.9); got != 5 {
		t.Errorf("FloorToInt(5.9) = %d, want 5", got)
	}
	if got := Math.FloorToInt(-5.9); got != -6 {
		t.Errorf("FloorToInt(-5.9) = %d, want -6（负数朝负无穷取整）", got)
	}
}

func TestMathRoundToIntAndFloat(t *testing.T) {
	if got := Math.RoundToInt(-5.5); got != -6 {
		t.Errorf("RoundToInt(-5.5) = %d, want -6（.5 远离零）", got)
	}
	if got := Math.RoundToInt(5.4); got != 5 {
		t.Errorf("RoundToInt(5.4) = %d, want 5（不足 .5 不进位）", got)
	}
	if got := Math.RoundFloat(1.23456, 4); got != 1.2346 {
		t.Errorf("RoundFloat(1.23456,4) = %v, want 1.2346", got)
	}
}

func TestMathPageCountEdges(t *testing.T) {
	if got := Math.PageCount(100, 20); got != 5 {
		t.Errorf("PageCount(100,20) 整除 = %d, want 5", got)
	}
	if got := Math.PageCount(1, 20); got != 1 {
		t.Errorf("PageCount(1,20) = %d, want 1", got)
	}
	if got := Math.PageCount(-5, 20); got != 0 {
		t.Errorf("PageCount(-5,20) = %d, want 0（total 非正）", got)
	}
}
