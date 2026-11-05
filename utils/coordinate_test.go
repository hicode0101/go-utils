package utils

import "testing"

// coordAbsF 求浮点数绝对值（原 utils_test.go 中 TestZipCoord 的 absF，
// 改名 coordAbsF 以免与其它测试文件的 absF 冲突）。
func coordAbsF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// TestCoordOutOfChina OutOfChina：境内点返回 false，境外点返回 true（含边界值）。
func TestCoordOutOfChina(t *testing.T) {
	cases := []struct {
		name       string
		lng, lat   float64
		wantOffset bool
	}{
		{"北京", 116.404, 39.915, false},
		{"上海", 121.4737, 31.2304, false},
		{"西边界线上", 72.004, 35.0, false},
		{"东边界线上", 137.8347, 35.0, false},
		{"原点(大西洋)", 0, 0, true},
		{"纽约", -74.006, 40.713, true},
		{"经度越东界", 140.0, 35.0, true},
		{"纬度越北界", 116.0, 60.0, true},
		{"纬度越南界", 116.0, 0.5, true},
	}
	for _, c := range cases {
		if got := Coord.OutOfChina(c.lng, c.lat); got != c.wantOffset {
			t.Errorf("%s OutOfChina(%v,%v) = %v, want %v", c.name, c.lng, c.lat, got, c.wantOffset)
		}
	}
}

// TestCoordWgs84ToGcj02 Wgs84ToGcj02：hutool 官方黄金值 + 偏移量合理范围 + 境外原样返回。
func TestCoordWgs84ToGcj02(t *testing.T) {
	// 黄金值对照 hutool CoordinateUtilTest#wgs84ToGcj02Test
	gcj := Coord.Wgs84ToGcj02(116.404, 39.915)
	if coordAbsF(gcj.Lng-116.41024449916938) > 1e-9 || coordAbsF(gcj.Lat-39.91640428150164) > 1e-9 {
		t.Errorf("Wgs84ToGcj02(北京) = %v, %v; want 116.41024449916938, 39.91640428150164", gcj.Lng, gcj.Lat)
	}
	// 偏移量应存在且在合理范围（火星坐标偏移通常在千分之几度以内）
	const lng, lat = 116.404, 39.915
	dlng, dlat := gcj.Lng-lng, gcj.Lat-lat
	if dlng <= 0 || dlat <= 0 {
		t.Errorf("GCJ-02 偏移应为正: dlng=%v, dlat=%v", dlng, dlat)
	}
	if dlng >= 0.01 || dlat >= 0.01 {
		t.Errorf("GCJ-02 偏移量过大: dlng=%v, dlat=%v", dlng, dlat)
	}
	// 境外坐标原样返回
	for _, p := range [][2]float64{{0, 0}, {-74.006, 40.713}, {139.69, 35.68}} {
		if got := Coord.Wgs84ToGcj02(p[0], p[1]); got.Lng != p[0] || got.Lat != p[1] {
			t.Errorf("境外点 %v 转换应原样返回, got %v", p, got)
		}
	}
}

// TestCoordGcj02ToWgs84 Gcj02ToWgs84：hutool 官方黄金值 + 往返误差 < 0.002 + 境外原样返回。
func TestCoordGcj02ToWgs84(t *testing.T) {
	// 黄金值对照 hutool CoordinateUtilTest#gcj02ToWgs84Test
	wgs := Coord.Gcj02ToWgs84(116.404, 39.915)
	if coordAbsF(wgs.Lng-116.39775550083061) > 1e-9 || coordAbsF(wgs.Lat-39.91359571849836) > 1e-9 {
		t.Errorf("Gcj02ToWgs84(北京) = %v, %v; want 116.39775550083061, 39.91359571849836", wgs.Lng, wgs.Lat)
	}
	// WGS84 → GCJ-02 → WGS84 往返误差应在米级（< 0.002 度）以内
	const lng, lat = 116.404, 39.915
	gcj := Coord.Wgs84ToGcj02(lng, lat)
	back := Coord.Gcj02ToWgs84(gcj.Lng, gcj.Lat)
	if coordAbsF(back.Lng-lng) > 0.002 || coordAbsF(back.Lat-lat) > 0.002 {
		t.Errorf("GCJ-02 往返误差过大: %v, %v", back.Lng, back.Lat)
	}
	// 偏移应存在且方向为负
	if back.Lng >= gcj.Lng || back.Lat >= gcj.Lat {
		t.Errorf("Gcj02ToWgs84 应产生负向偏移: %v -> %v", gcj, back)
	}
	// 境外坐标原样返回
	if got := Coord.Gcj02ToWgs84(0, 0); got.Lng != 0 || got.Lat != 0 {
		t.Errorf("境外点应原样返回, got %v", got)
	}
}

// TestCoordGcj02Bd09 Gcj02ToBd09/Bd09ToGcj02：hutool 官方黄金值 + 偏移存在 + 往返误差小。
func TestCoordGcj02Bd09(t *testing.T) {
	const lng, lat = 116.404, 39.915
	// 黄金值对照 hutool CoordinateUtilTest#gcj02ToBd09Test
	bd := Coord.Gcj02ToBd09(lng, lat)
	if coordAbsF(bd.Lng-116.41036949371029) > 1e-9 || coordAbsF(bd.Lat-39.92133699351022) > 1e-9 {
		t.Errorf("Gcj02ToBd09(北京) = %v, %v; want 116.41036949371029, 39.92133699351022", bd.Lng, bd.Lat)
	}
	// 黄金值对照 hutool CoordinateUtilTest#bd09toGcj02Test
	gcj := Coord.Bd09ToGcj02(lng, lat)
	if coordAbsF(gcj.Lng-116.39762729119315) > 1e-9 || coordAbsF(gcj.Lat-39.90865673957631) > 1e-9 {
		t.Errorf("Bd09ToGcj02(北京) = %v, %v; want 116.39762729119315, 39.90865673957631", gcj.Lng, gcj.Lat)
	}
	// BD-09 相对 GCJ-02 应存在偏移
	if bd.Lng == lng && bd.Lat == lat {
		t.Error("BD-09 转换应产生偏移")
	}
	// 往返误差应远小于一个偏移量级
	back := Coord.Bd09ToGcj02(bd.Lng, bd.Lat)
	if coordAbsF(back.Lng-lng) > 1e-4 || coordAbsF(back.Lat-lat) > 1e-4 {
		t.Errorf("GCJ-02/BD-09 往返误差过大: %v, %v", back.Lng, back.Lat)
	}
	back2 := Coord.Gcj02ToBd09(gcj.Lng, gcj.Lat)
	if coordAbsF(back2.Lng-lng) > 1e-4 || coordAbsF(back2.Lat-lat) > 1e-4 {
		t.Errorf("BD-09/GCJ-02 往返误差过大: %v, %v", back2.Lng, back2.Lat)
	}
}

// TestCoordWgs84Bd09 Wgs84ToBd09/Bd09ToWgs84：hutool 官方黄金值（含第二组远端点）+ 往返。
func TestCoordWgs84Bd09(t *testing.T) {
	const lng, lat = 116.404, 39.915
	// 黄金值对照 hutool CoordinateUtilTest#wgs84toBd09Test
	bd := Coord.Wgs84ToBd09(lng, lat)
	if coordAbsF(bd.Lng-116.41662724378733) > 1e-9 || coordAbsF(bd.Lat-39.922699552216216) > 1e-9 {
		t.Errorf("Wgs84ToBd09(北京) = %v, %v; want 116.41662724378733, 39.922699552216216", bd.Lng, bd.Lat)
	}
	// 黄金值对照 hutool CoordinateUtilTest#wgs84toBd09Test2（https://tool.lu/coordinate/）
	bd2 := Coord.Wgs84ToBd09(122.99395597, 44.99804071)
	if coordAbsF(bd2.Lng-123.00636516028885) > 1e-9 || coordAbsF(bd2.Lat-45.00636909189589) > 1e-9 {
		t.Errorf("Wgs84ToBd09(内蒙) = %v, %v; want 123.00636516028885, 45.00636909189589", bd2.Lng, bd2.Lat)
	}
	// 黄金值对照 hutool CoordinateUtilTest#bd09toWgs84Test
	wgs := Coord.Bd09ToWgs84(lng, lat)
	if coordAbsF(wgs.Lng-116.3913836995125) > 1e-9 || coordAbsF(wgs.Lat-39.907253214522164) > 1e-9 {
		t.Errorf("Bd09ToWgs84(北京) = %v, %v; want 116.3913836995125, 39.907253214522164", wgs.Lng, wgs.Lat)
	}
	// WGS84 → BD-09 → WGS84 往返误差应在米级以内
	back := Coord.Bd09ToWgs84(bd.Lng, bd.Lat)
	if coordAbsF(back.Lng-lng) > 0.002 || coordAbsF(back.Lat-lat) > 0.002 {
		t.Errorf("WGS84/BD-09 往返误差过大: %v, %v", back.Lng, back.Lat)
	}
}

// TestCoordMercator Wgs84ToMercator/MercatorToWgs84：验证过的黄金值 + 往返误差极小 + 投影边界。
func TestCoordMercator(t *testing.T) {
	// 北京 (116.404, 39.915) 的 Web 墨卡托投影（先用独立程序验证后固化）：
	// X ≈ 12958034.007 米（东向），Y ≈ 4853597.988 米（北向）
	m := Coord.Wgs84ToMercator(116.404, 39.915)
	if coordAbsF(m.Lng-12958034.006300217) > 0.01 || coordAbsF(m.Lat-4853597.98829984) > 0.01 {
		t.Errorf("Wgs84ToMercator(北京) = %v, %v; want 约 12958034.006, 4853597.988", m.Lng, m.Lat)
	}
	// 逆投影还原经纬度
	back := Coord.MercatorToWgs84(m.Lng, m.Lat)
	if coordAbsF(back.Lng-116.404) > 1e-6 || coordAbsF(back.Lat-39.915) > 1e-6 {
		t.Errorf("MercatorToWgs84 逆投影 = %v, %v; want 116.404, 39.915", back.Lng, back.Lat)
	}
	// 任意点往返误差 < 0.01
	for _, p := range [][2]float64{{0, 0}, {121.4737, 31.2304}, {-74.006, 40.713}, {89.5, -23.4}} {
		proj := Coord.Wgs84ToMercator(p[0], p[1])
		rt := Coord.MercatorToWgs84(proj.Lng, proj.Lat)
		if coordAbsF(rt.Lng-p[0]) > 0.01 || coordAbsF(rt.Lat-p[1]) > 0.01 {
			t.Errorf("墨卡托往返 %v 得到 %v，误差过大", p, rt)
		}
	}
	// 投影边界：经度 ±180 对应 X = ±coordMercatorAxis
	if got := Coord.Wgs84ToMercator(180, 0).Lng; coordAbsF(got-20037508.342789244) > 1e-6 {
		t.Errorf("经度 180 的投影 X = %v, want 20037508.342789244", got)
	}
	if got := Coord.MercatorToWgs84(20037508.342789244, 0).Lng; coordAbsF(got-180) > 1e-6 {
		t.Errorf("投影 X=axis 反算经度 = %v, want 180", got)
	}
	if got := Coord.Wgs84ToMercator(0, 0); got.Lng != 0 || got.Lat != 0 {
		t.Errorf("(0,0) 投影应为原点, got %v", got)
	}
}
