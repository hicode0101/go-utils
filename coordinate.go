package utils

import "math"

// utilCoord 坐标转换工具。
//
// 主流坐标系说明（对照 hutool CoordinateUtil）：
//   - WGS84：地球坐标系，GPS 设备的原始坐标，中国以外地区谷歌地图使用；
//   - GCJ-02：火星坐标系（国测局加密坐标），高德、腾讯、阿里等国内地图使用；
//   - BD-09：百度坐标系，在 GCJ-02 基础上二次加密，百度、搜狗等使用；
//   - Web 墨卡托：Web 地图通用的平面投影坐标（单位：米）。
//
// 坐标转换参考: https://tool.lu/coordinate/
type utilCoord struct{}

// Coordinate 经纬度坐标（Lng 经度，Lat 纬度），对照 hutool 的 CoordinateUtil.Coordinate。
// 因包内已存在工具单例变量 utils.Coord，而 Go 不允许包级变量与类型同名，故类型命名为 Coordinate。
type Coordinate struct {
	Lng float64 // 经度（longitude）
	Lat float64 // 纬度（latitude）
}

// 坐标转换常量（取值与公式逐项对照 hutool CoordinateUtil）。
const (
	// coordXPi 火星坐标系与百度坐标系转换的中间量：π * 3000 / 180 ≈ 52.35987755982988
	coordXPi = 3.1415926535897932384626433832795 * 3000.0 / 180.0
	// coordPi 圆周率 π
	coordPi = 3.1415926535897932384626433832795
	// coordRadius 地球半径（Krasovsky 1940 椭球长半轴，米），对应 hutool RADIUS
	coordRadius = 6378245.0
	// coordCorrectionParam 椭球偏率 ee，对应 hutool CORRECTION_PARAM
	coordCorrectionParam = 0.00669342162296594323
	// coordMercatorAxis Web 墨卡托投影赤道半长（米），即经度 ±180° 对应的 X 值
	coordMercatorAxis = 20037508.342789244
)

/* ---------------- 判断 ---------------- */

// OutOfChina 判断经纬度是否在中国境外。
// GCJ-02 偏移算法只对国内坐标有效，境外坐标应原样使用（各转换方法内部已做此特判）。
// 场景：过滤境外 GPS 点位，跳过无意义的偏移换算。
func (c *utilCoord) OutOfChina(lng, lat float64) bool {
	return (lng < 72.004 || lng > 137.8347) || (lat < 0.8293 || lat > 55.8271)
}

/* ---------------- WGS84 ---------------- */

// Wgs84ToGcj02 将 WGS84 坐标（GPS 原始坐标）转换为火星坐标系 (GCJ-02)。
// 场景：把车载 GPS 上报的坐标显示到高德/腾讯地图上。
// 中国境外坐标不做偏移，原样返回（对应 hutool outOfChina 特判）。
func (c *utilCoord) Wgs84ToGcj02(lng, lat float64) Coordinate {
	if c.OutOfChina(lng, lat) {
		return Coordinate{Lng: lng, Lat: lat}
	}
	offset := coordOffset(lng, lat, true)
	return Coordinate{Lng: lng + offset.Lng, Lat: lat + offset.Lat}
}

// Wgs84ToBd09 将 WGS84 坐标转换为百度坐标系 (BD-09)。
// 内部先转 GCJ-02 再转 BD-09（对照 hutool wgs84ToBd09）。
// 场景：把 GPS 轨迹点标注到百度地图上。
func (c *utilCoord) Wgs84ToBd09(lng, lat float64) Coordinate {
	gcj := c.Wgs84ToGcj02(lng, lat)
	return c.Gcj02ToBd09(gcj.Lng, gcj.Lat)
}

/* ---------------- GCJ-02 ---------------- */

// Gcj02ToWgs84 将火星坐标系 (GCJ-02) 转换回 WGS84。
// 场景：把高德地图的选点结果还原成 GPS 原始坐标。
// 中国境外坐标不做偏移，原样返回（对应 hutool outOfChina 特判）。
func (c *utilCoord) Gcj02ToWgs84(lng, lat float64) Coordinate {
	if c.OutOfChina(lng, lat) {
		return Coordinate{Lng: lng, Lat: lat}
	}
	offset := coordOffset(lng, lat, false)
	return Coordinate{Lng: lng + offset.Lng, Lat: lat + offset.Lat}
}

// Gcj02ToBd09 将火星坐标系 (GCJ-02) 转换为百度坐标系 (BD-09)。
// 场景：高德坐标迁移到百度地图标注。
func (c *utilCoord) Gcj02ToBd09(lng, lat float64) Coordinate {
	z := math.Sqrt(lng*lng+lat*lat) + 0.00002*math.Sin(lat*coordXPi)
	theta := math.Atan2(lat, lng) + 0.000003*math.Cos(lng*coordXPi)
	return Coordinate{
		Lng: z*math.Cos(theta) + 0.0065,
		Lat: z*math.Sin(theta) + 0.006,
	}
}

/* ---------------- BD-09 ---------------- */

// Bd09ToGcj02 将百度坐标系 (BD-09) 转换为火星坐标系 (GCJ-02)，即百度转谷歌、高德。
func (c *utilCoord) Bd09ToGcj02(lng, lat float64) Coordinate {
	x := lng - 0.0065
	y := lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*coordXPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*coordXPi)
	return Coordinate{Lng: z * math.Cos(theta), Lat: z * math.Sin(theta)}
}

// Bd09ToWgs84 将百度坐标系 (BD-09) 转换为 WGS84。
// 内部先转 GCJ-02 再转 WGS84（对照 hutool bd09toWgs84）。
// 场景：把百度地图采集的点还原成 GPS 原始坐标。
func (c *utilCoord) Bd09ToWgs84(lng, lat float64) Coordinate {
	gcj := c.Bd09ToGcj02(lng, lat)
	return c.Gcj02ToWgs84(gcj.Lng, gcj.Lat)
}

/* ---------------- 墨卡托投影 ---------------- */

// Wgs84ToMercator 将 WGS84 经纬度转换为 Web 墨卡托投影坐标（单位：米）。
// 转换结果 Lng 存放投影 X（东向），Lat 存放投影 Y（北向）。
// 场景：切片地图服务的瓦片坐标计算。
func (c *utilCoord) Wgs84ToMercator(lng, lat float64) Coordinate {
	x := lng * coordMercatorAxis / 180
	y := math.Log(math.Tan((90+lat)*coordPi/360)) / (coordPi / 180)
	y = y * coordMercatorAxis / 180
	return Coordinate{Lng: x, Lat: y}
}

// MercatorToWgs84 将 Web 墨卡托投影坐标转换回 WGS84 经纬度。
// 入参 mercatorX/mercatorY 分别为投影坐标的 X/Y（米）。
func (c *utilCoord) MercatorToWgs84(mercatorX, mercatorY float64) Coordinate {
	x := mercatorX / coordMercatorAxis * 180
	y := mercatorY / coordMercatorAxis * 180
	y = 180 / coordPi * (2*math.Atan(math.Exp(y*coordPi/180)) - coordPi/2)
	return Coordinate{Lng: x, Lat: y}
}

/* ---------------- 内部算法 ---------------- */

// coordOffset 计算 WGS84 与 GCJ-02 互转的偏移量（非精确算法，对照 hutool offset）。
// isPlus 为 true 表示正向偏移（WGS84→GCJ-02），false 表示反向（GCJ-02→WGS84）。
func coordOffset(lng, lat float64, isPlus bool) Coordinate {
	dlng := coordTransLng(lng-105.0, lat-35.0)
	dlat := coordTransLat(lng-105.0, lat-35.0)

	magic := math.Sin(lat / 180.0 * coordPi)
	magic = 1 - coordCorrectionParam*magic*magic
	sqrtMagic := math.Sqrt(magic)

	dlng = dlng * 180.0 / (coordRadius / sqrtMagic * math.Cos(lat/180.0*coordPi) * coordPi)
	dlat = dlat * 180.0 / (coordRadius * (1 - coordCorrectionParam) / (magic * sqrtMagic) * coordPi)

	if !isPlus {
		dlng = -dlng
		dlat = -dlat
	}
	return Coordinate{Lng: dlng, Lat: dlat}
}

// coordTransLng 偏移算法的经度分量计算（对照 hutool transLng）。
func coordTransLng(lng, lat float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*coordPi) + 20.0*math.Sin(2.0*lng*coordPi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*coordPi) + 40.0*math.Sin(lng/3.0*coordPi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*coordPi) + 300.0*math.Sin(lng/30.0*coordPi)) * 2.0 / 3.0
	return ret
}

// coordTransLat 偏移算法的纬度分量计算（对照 hutool transLat）。
func coordTransLat(lng, lat float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*coordPi) + 20.0*math.Sin(2.0*lng*coordPi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*coordPi) + 40.0*math.Sin(lat/3.0*coordPi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*coordPi) + 320.0*math.Sin(lat*coordPi/30.0)) * 2.0 / 3.0
	return ret
}
