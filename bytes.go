package utils

import (
	"bytes"
	"encoding/binary"
	"math"
	"unicode/utf16"
)

// utilBytes 字节流编解码工具（统一小端序 LittleEndian）。
//
// 典型场景：解析设备/硬件/网络协议的二进制定长报文。
//
// 值范围速查（1 byte = 8 bit）：
//
//	uint8: 0~255              int8: -128~127
//	uint16: 0~65535           int16: -32768~32767
//	uint32: 0~4294967295      int32: -2147483648~2147483647
//	uint64: 0~18446744073709551615          int64: -9223372036854775808~9223372036854775807
//	float32: 约 3.4E-38~3.4E+38             float64: 约 1.7E-308~1.7E+308
//
// 约定：写入方法的 buf 为 *bytes.Buffer（其 Write 永不返回错误，故不再返回 error）；
// 读取方法不做越界检查，输入切片长度不足时会 panic（与 binary.LittleEndian 一致），
// 调用方须保证切片长度（报文解析前先校验长度）。
type utilBytes struct{}

/* ---------------- 定长数值写入 ---------------- */

// WriteUint8 写入 1 字节无符号整数。场景：报文中的命令字/标志字段。
func (b *utilBytes) WriteUint8(buf *bytes.Buffer, val uint8) {
	buf.WriteByte(val)
}

// WriteUint16 写入 2 字节小端序无符号整数。
// 场景：报文长度字段。WriteUint16(buf, 0x1234) 产出字节 34 12。
func (b *utilBytes) WriteUint16(buf *bytes.Buffer, val uint16) {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, val)
	buf.Write(bs)
}

// WriteUint32 写入 4 字节小端序无符号整数。场景：IP 数值、CRC 校验值。
func (b *utilBytes) WriteUint32(buf *bytes.Buffer, val uint32) {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, val)
	buf.Write(bs)
}

// WriteUint64 写入 8 字节小端序无符号整数。场景：8 字节雪花 ID 序列化。
func (b *utilBytes) WriteUint64(buf *bytes.Buffer, val uint64) {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, val)
	buf.Write(bs)
}

// WriteInt8 写入 1 字节有符号整数。
func (b *utilBytes) WriteInt8(buf *bytes.Buffer, val int8) {
	buf.WriteByte(byte(val))
}

// WriteInt16 写入 2 字节小端序有符号整数。场景：温度等带符号测量值。
func (b *utilBytes) WriteInt16(buf *bytes.Buffer, val int16) {
	b.WriteUint16(buf, uint16(val))
}

// WriteInt32 写入 4 字节小端序有符号整数。场景：带符号坐标值。
func (b *utilBytes) WriteInt32(buf *bytes.Buffer, val int32) {
	b.WriteUint32(buf, uint32(val))
}

// WriteInt64 写入 8 字节小端序有符号整数。场景：毫秒时间戳序列化。
func (b *utilBytes) WriteInt64(buf *bytes.Buffer, val int64) {
	b.WriteUint64(buf, uint64(val))
}

// WriteFloat32 写入 4 字节小端序单精度浮点数。场景：传感器浮点测量值。
func (b *utilBytes) WriteFloat32(buf *bytes.Buffer, val float32) {
	b.WriteUint32(buf, math.Float32bits(val))
}

// WriteFloat64 写入 8 字节小端序双精度浮点数。场景：高精度坐标序列化。
func (b *utilBytes) WriteFloat64(buf *bytes.Buffer, val float64) {
	b.WriteUint64(buf, math.Float64bits(val))
}

/* ---------------- 定长数值读取 ---------------- */

// ReadUint8 读取 1 字节无符号整数（输入长度须 >= 1）。
func (b *utilBytes) ReadUint8(val []byte) uint8 {
	return val[0]
}

// ReadUint16 读取 2 字节小端序无符号整数（输入长度须 >= 2）。
func (b *utilBytes) ReadUint16(val []byte) uint16 {
	return binary.LittleEndian.Uint16(val)
}

// ReadUint32 读取 4 字节小端序无符号整数（输入长度须 >= 4）。
func (b *utilBytes) ReadUint32(val []byte) uint32 {
	return binary.LittleEndian.Uint32(val)
}

// ReadUint64 读取 8 字节小端序无符号整数（输入长度须 >= 8）。
func (b *utilBytes) ReadUint64(val []byte) uint64 {
	return binary.LittleEndian.Uint64(val)
}

// ReadInt8 读取 1 字节有符号整数（输入长度须 >= 1）。
func (b *utilBytes) ReadInt8(val []byte) int8 {
	return int8(val[0])
}

// ReadInt16 读取 2 字节小端序有符号整数（输入长度须 >= 2）。
func (b *utilBytes) ReadInt16(val []byte) int16 {
	return int16(binary.LittleEndian.Uint16(val))
}

// ReadInt32 读取 4 字节小端序有符号整数（输入长度须 >= 4）。
func (b *utilBytes) ReadInt32(val []byte) int32 {
	return int32(binary.LittleEndian.Uint32(val))
}

// ReadInt64 读取 8 字节小端序有符号整数（输入长度须 >= 8）。
func (b *utilBytes) ReadInt64(val []byte) int64 {
	return int64(binary.LittleEndian.Uint64(val))
}

// ReadFloat32 读取 4 字节小端序单精度浮点数（输入长度须 >= 4）。
func (b *utilBytes) ReadFloat32(val []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(val))
}

// ReadFloat64 读取 8 字节小端序双精度浮点数（输入长度须 >= 8）。
func (b *utilBytes) ReadFloat64(val []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(val))
}

/* ---------------- 结构体整体编解码 ---------------- */

// ReadAny 用 binary.Read 按小端序把字节流解码到 result
// （result 须为定长类型或不含 padding/指针的定长结构体指针）。
// 场景：整段报文头一次性解码：ReadAny(head, &msgHead)。
func (b *utilBytes) ReadAny(val []byte, result interface{}) error {
	return binary.Read(bytes.NewReader(val), binary.LittleEndian, result)
}

// WriteAny 用 binary.Write 按小端序把定长数据（或定长结构体）写入 buf。
// 场景：整段报文头一次性编码。
func (b *utilBytes) WriteAny(buf *bytes.Buffer, data interface{}) error {
	return binary.Write(buf, binary.LittleEndian, data)
}

/* ---------------- C 风格定长字符串（char[size] / WCHAR[size]） ---------------- */

// WriteFixedString 写入定长字符串：超长截断，不足右侧补 0（对应 C 的 char[size]）。
// WriteFixedString(buf, 8, "ab") 产出 61 62 00 00 00 00 00 00。
// 场景：与 C/C++ 设备程序的报文交互。
func (b *utilBytes) WriteFixedString(buf *bytes.Buffer, size int, val string) {
	bs := []byte(val)
	if len(bs) > size {
		bs = bs[:size]
	}
	buf.Write(bs)
	buf.Write(make([]byte, size-len(bs)))
}

// ReadFixedString 读取定长字符串并去掉末尾补位的 0。
// 场景：解析 C 报文中的定长名称字段。
func (b *utilBytes) ReadFixedString(val []byte) string {
	return string(bytes.TrimRight(val, "\x00"))
}

// WriteUtf16String 写入 UTF-16LE 定长字符串（对应 Windows TCHAR/WCHAR 宽字符，
// size 为字符数而非字节数，实际占用 size*2 字节），超长截断，不足补 0。
// WriteUtf16String(buf, 4, "你好") 产出 fd 59 7d 4f 00 00 00 00。
// 场景：与 Windows 程序/中文工控设备的宽字符报文交互。
func (b *utilBytes) WriteUtf16String(buf *bytes.Buffer, size int, val string) {
	units := utf16.Encode([]rune(val))
	if len(units) > size {
		units = units[:size]
	}
	bs := make([]byte, size*2)
	for i, u := range units {
		bs[i*2] = byte(u)
		bs[i*2+1] = byte(u >> 8)
	}
	buf.Write(bs)
}

// ReadUtf16String 把字节流按 UTF-16LE 解码为字符串（自动去掉末尾的 0 字符）。
// 场景：解析宽字符报文。
func (b *utilBytes) ReadUtf16String(val []byte) string {
	units := make([]uint16, 0, len(val)/2)
	for i := 0; i+1 < len(val); i += 2 {
		units = append(units, binary.LittleEndian.Uint16(val[i:]))
	}
	runes := utf16.Decode(units)
	end := len(runes)
	for end > 0 && runes[end-1] == 0 {
		end--
	}
	return string(runes[:end])
}
