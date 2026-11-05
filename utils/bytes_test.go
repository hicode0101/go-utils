package utils

import (
	"bytes"
	"testing"
)

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

func TestBytes(t *testing.T) {
	var buf bytes.Buffer
	Bytes.WriteUint16(&buf, 0x1234)
	if got := Bytes.ReadUint16(buf.Bytes()); got != 0x1234 {
		t.Errorf("Uint16 往返 = %#x", got)
	}

	buf.Reset()
	Bytes.WriteInt64(&buf, -99)
	if got := Bytes.ReadInt64(buf.Bytes()); got != -99 {
		t.Errorf("Int64 往返 = %d", got)
	}

	buf.Reset()
	Bytes.WriteFloat64(&buf, 3.14)
	if got := Bytes.ReadFloat64(buf.Bytes()); got != 3.14 {
		t.Errorf("Float64 往返 = %v", got)
	}

	buf.Reset()
	Bytes.WriteFixedString(&buf, 8, "ab")
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "ab" {
		t.Errorf("FixedString 往返 = %q, want \"ab\"", got)
	}
	if buf.Len() != 8 {
		t.Errorf("定长字符串应占 8 字节，实际 %d", buf.Len())
	}

	buf.Reset()
	Bytes.WriteUtf16String(&buf, 4, "你好")
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "你好" {
		t.Errorf("Utf16String 往返 = %q, want \"你好\"", got)
	}
	if buf.Len() != 8 { // 4 字符 * 2 字节
		t.Errorf("UTF-16 定长应占 8 字节，实际 %d", buf.Len())
	}

	// 超长截断
	buf.Reset()
	Bytes.WriteFixedString(&buf, 3, "hello")
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "hel" {
		t.Errorf("超长截断 = %q, want \"hel\"", got)
	}

	// 大端序互转（hutool ByteUtil）
	be := Bytes.Int32ToBytes(0x01020304, true)
	if got := Bytes.BytesToInt32(be, true); got != 0x01020304 {
		t.Errorf("大端 Int32 往返 = %#x", got)
	}
	if be[0] != 0x01 {
		t.Errorf("大端序首字节应为 0x01, got %#x", be[0])
	}
	le := Bytes.Int64ToBytes(-2, false)
	if got := Bytes.BytesToInt64(le, false); got != -2 {
		t.Errorf("小端 Int64 往返 = %d", got)
	}
	if got := Bytes.BytesToFloat64(Bytes.Float64ToBytes(2.5, true), true); got != 2.5 {
		t.Errorf("大端 Float64 往返 = %v", got)
	}
	if got := Bytes.NumberToBytes(int32(0x11223344), false); got == nil || len(got) != 4 {
		t.Errorf("NumberToBytes 分派失败: %v", got)
	}
}

/* ---------------- 无符号整数读写 ---------------- */

func TestBytesUintRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	// uint8：边界 0 与最大值 255
	for _, v := range []uint8{0, 1, 255} {
		buf.Reset()
		Bytes.WriteUint8(&buf, v)
		if buf.Len() != 1 {
			t.Errorf("WriteUint8 应写 1 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadUint8(buf.Bytes()); got != v {
			t.Errorf("uint8 往返 = %d, want %d", got, v)
		}
	}
	// uint16
	for _, v := range []uint16{0, 0x1234, 0xFFFF} {
		buf.Reset()
		Bytes.WriteUint16(&buf, v)
		if buf.Len() != 2 {
			t.Errorf("WriteUint16 应写 2 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadUint16(buf.Bytes()); got != v {
			t.Errorf("uint16 往返 = %d, want %d", got, v)
		}
	}
	// 小端序字节序断言：0x1234 产出 34 12
	buf.Reset()
	Bytes.WriteUint16(&buf, 0x1234)
	if bs := buf.Bytes(); bs[0] != 0x34 || bs[1] != 0x12 {
		t.Errorf("WriteUint16 小端序错误: % x, want 34 12", bs)
	}
	// uint32
	for _, v := range []uint32{0, 0x12345678, 0xFFFFFFFF} {
		buf.Reset()
		Bytes.WriteUint32(&buf, v)
		if buf.Len() != 4 {
			t.Errorf("WriteUint32 应写 4 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadUint32(buf.Bytes()); got != v {
			t.Errorf("uint32 往返 = %d, want %d", got, v)
		}
	}
	buf.Reset()
	Bytes.WriteUint32(&buf, 0x12345678)
	if bs := buf.Bytes(); bs[0] != 0x78 {
		t.Errorf("WriteUint32 小端序首字节错误: % x", bs)
	}
	// uint64
	for _, v := range []uint64{0, 0x0123456789ABCDEF, 0xFFFFFFFFFFFFFFFF} {
		buf.Reset()
		Bytes.WriteUint64(&buf, v)
		if buf.Len() != 8 {
			t.Errorf("WriteUint64 应写 8 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadUint64(buf.Bytes()); got != v {
			t.Errorf("uint64 往返 = %d, want %d", got, v)
		}
	}
	buf.Reset()
	Bytes.WriteUint64(&buf, 0x0123456789ABCDEF)
	if bs := buf.Bytes(); bs[0] != 0xEF {
		t.Errorf("WriteUint64 小端序首字节错误: % x", bs)
	}
}

/* ---------------- 有符号整数读写（含负数与 0） ---------------- */

func TestBytesIntRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	for _, v := range []int8{-128, -1, 0, 127} {
		buf.Reset()
		Bytes.WriteInt8(&buf, v)
		if buf.Len() != 1 {
			t.Errorf("WriteInt8 应写 1 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadInt8(buf.Bytes()); got != v {
			t.Errorf("int8 往返 = %d, want %d", got, v)
		}
	}
	for _, v := range []int16{-32768, -1, 0, 32767} {
		buf.Reset()
		Bytes.WriteInt16(&buf, v)
		if buf.Len() != 2 {
			t.Errorf("WriteInt16 应写 2 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadInt16(buf.Bytes()); got != v {
			t.Errorf("int16 往返 = %d, want %d", got, v)
		}
	}
	for _, v := range []int32{-2147483648, -1, 0, 2147483647} {
		buf.Reset()
		Bytes.WriteInt32(&buf, v)
		if buf.Len() != 4 {
			t.Errorf("WriteInt32 应写 4 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadInt32(buf.Bytes()); got != v {
			t.Errorf("int32 往返 = %d, want %d", got, v)
		}
	}
	for _, v := range []int64{-9223372036854775808, -99, 0, 9223372036854775807} {
		buf.Reset()
		Bytes.WriteInt64(&buf, v)
		if buf.Len() != 8 {
			t.Errorf("WriteInt64 应写 8 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadInt64(buf.Bytes()); got != v {
			t.Errorf("int64 往返 = %d, want %d", got, v)
		}
	}
	// 负数小端字节序：int16(-2) = 0xFFFE → FE FF
	buf.Reset()
	Bytes.WriteInt16(&buf, -2)
	if bs := buf.Bytes(); bs[0] != 0xFE || bs[1] != 0xFF {
		t.Errorf("WriteInt16(-2) 字节序错误: % x, want fe ff", bs)
	}
}

/* ---------------- 浮点读写（含负数与 0） ---------------- */

func TestBytesFloatRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	for _, v := range []float32{0, -3.5, 1.25, 3.4e38} {
		buf.Reset()
		Bytes.WriteFloat32(&buf, v)
		if buf.Len() != 4 {
			t.Errorf("WriteFloat32 应写 4 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadFloat32(buf.Bytes()); got != v {
			t.Errorf("float32 往返 = %v, want %v", got, v)
		}
	}
	for _, v := range []float64{0, -3.14, 2.5, 1.7e308} {
		buf.Reset()
		Bytes.WriteFloat64(&buf, v)
		if buf.Len() != 8 {
			t.Errorf("WriteFloat64 应写 8 字节, got %d", buf.Len())
		}
		if got := Bytes.ReadFloat64(buf.Bytes()); got != v {
			t.Errorf("float64 往返 = %v, want %v", got, v)
		}
	}
}

/* ---------------- 整数与字节切片互转（大小端两种） ---------------- */

func TestBytesEndianInt(t *testing.T) {
	// 大端首字节为高位，小端首字节为低位
	if be := Bytes.Int16ToBytes(0x1234, true); be[0] != 0x12 || be[1] != 0x34 {
		t.Errorf("Int16ToBytes 大端 = % x, want 12 34", be)
	}
	if le := Bytes.Int16ToBytes(0x1234, false); le[0] != 0x34 || le[1] != 0x12 {
		t.Errorf("Int16ToBytes 小端 = % x, want 34 12", le)
	}
	if got := Bytes.BytesToInt16(Bytes.Int16ToBytes(-2, true), true); got != -2 {
		t.Errorf("大端 int16 负数往返 = %d, want -2", got)
	}
	if got := Bytes.BytesToInt16(Bytes.Int16ToBytes(-2, false), false); got != -2 {
		t.Errorf("小端 int16 负数往返 = %d, want -2", got)
	}

	if be := Bytes.Int32ToBytes(0x01020304, true); be[0] != 0x01 {
		t.Errorf("Int32ToBytes 大端首字节 = %#x, want 0x01", be[0])
	}
	if le := Bytes.Int32ToBytes(0x01020304, false); le[0] != 0x04 {
		t.Errorf("Int32ToBytes 小端首字节 = %#x, want 0x04", le[0])
	}
	if got := Bytes.BytesToInt32(Bytes.Int32ToBytes(0x01020304, true), true); got != 0x01020304 {
		t.Errorf("大端 int32 往返 = %#x", got)
	}
	if got := Bytes.BytesToInt32(Bytes.Int32ToBytes(-1, false), false); got != -1 {
		t.Errorf("小端 int32 负数往返 = %d", got)
	}

	if be := Bytes.Int64ToBytes(0x0102030405060708, true); be[0] != 0x01 {
		t.Errorf("Int64ToBytes 大端首字节 = %#x, want 0x01", be[0])
	}
	if le := Bytes.Int64ToBytes(0x0102030405060708, false); le[0] != 0x08 {
		t.Errorf("Int64ToBytes 小端首字节 = %#x, want 0x08", le[0])
	}
	if got := Bytes.BytesToInt64(Bytes.Int64ToBytes(-2, true), true); got != -2 {
		t.Errorf("大端 int64 负数往返 = %d", got)
	}
	if got := Bytes.BytesToInt64(Bytes.Int64ToBytes(0x0102030405060708, false), false); got != 0x0102030405060708 {
		t.Errorf("小端 int64 往返 = %#x", got)
	}
}

/* ---------------- 浮点与字节切片互转（大小端两种） ---------------- */

func TestBytesEndianFloat(t *testing.T) {
	for _, bigEndian := range []bool{true, false} {
		if got := Bytes.BytesToFloat32(Bytes.Float32ToBytes(1.5, bigEndian), bigEndian); got != 1.5 {
			t.Errorf("Float32 往返(endian=%v) = %v, want 1.5", bigEndian, got)
		}
		if got := Bytes.BytesToFloat32(Bytes.Float32ToBytes(0, bigEndian), bigEndian); got != 0 {
			t.Errorf("Float32(0) 往返(endian=%v) = %v", bigEndian, got)
		}
		if got := Bytes.BytesToFloat64(Bytes.Float64ToBytes(-3.14, bigEndian), bigEndian); got != -3.14 {
			t.Errorf("Float64 往返(endian=%v) = %v, want -3.14", bigEndian, got)
		}
		if got := Bytes.BytesToFloat64(Bytes.Float64ToBytes(0, bigEndian), bigEndian); got != 0 {
			t.Errorf("Float64(0) 往返(endian=%v) = %v", bigEndian, got)
		}
	}
	// 1.5 = 0x3FC00000：大端首字节 0x3F，小端末字节 0x3F
	be := Bytes.Float32ToBytes(1.5, true)
	if be[0] != 0x3F || be[3] != 0x00 {
		t.Errorf("Float32ToBytes 大端 = % x, want 3f c0 00 00", be)
	}
	le := Bytes.Float32ToBytes(1.5, false)
	if le[0] != 0x00 || le[3] != 0x3F {
		t.Errorf("Float32ToBytes 小端 = % x, want 00 00 c0 3f", le)
	}
}

/* ---------------- 结构体整体编解码 ---------------- */

func TestBytesAny(t *testing.T) {
	type bytesMsgHead struct {
		Magic  uint16
		Length int32
		Score  float64
	}
	want := bytesMsgHead{Magic: 0xA55A, Length: -100, Score: 3.14}
	var buf bytes.Buffer
	if err := Bytes.WriteAny(&buf, want); err != nil {
		t.Fatalf("WriteAny 失败: %v", err)
	}
	if buf.Len() != 14 { // 2 + 4 + 8，binary 包按字段紧凑排列
		t.Errorf("定长结构体应占 14 字节, got %d", buf.Len())
	}
	var got bytesMsgHead
	if err := Bytes.ReadAny(buf.Bytes(), &got); err != nil {
		t.Fatalf("ReadAny 失败: %v", err)
	}
	if got != want {
		t.Errorf("结构体往返 = %+v, want %+v", got, want)
	}
	// 错误路径：输入长度不足
	if err := Bytes.ReadAny(buf.Bytes()[:5], &got); err == nil {
		t.Error("ReadAny 输入长度不足应返回错误")
	}
	// 错误路径：string 非定长类型不支持
	if err := Bytes.WriteAny(&buf, "abc"); err == nil {
		t.Error("WriteAny 写 string 应返回错误")
	}
}

/* ---------------- C 风格定长字符串 ---------------- */

func TestBytesFixedString(t *testing.T) {
	var buf bytes.Buffer
	// 不足补零
	Bytes.WriteFixedString(&buf, 8, "ab")
	if buf.Len() != 8 {
		t.Errorf("定长应占 8 字节, got %d", buf.Len())
	}
	bs := buf.Bytes()
	if bs[0] != 'a' || bs[1] != 'b' {
		t.Errorf("前两字节应为 'a','b': % x", bs)
	}
	for i := 2; i < 8; i++ {
		if bs[i] != 0x00 {
			t.Fatalf("右侧应补 0: % x", bs)
		}
	}
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "ab" {
		t.Errorf("ReadFixedString = %q, want \"ab\"", got)
	}
	// 超长截断
	buf.Reset()
	Bytes.WriteFixedString(&buf, 3, "hello")
	if buf.Len() != 3 {
		t.Errorf("截断后应占 3 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "hel" {
		t.Errorf("截断往返 = %q, want \"hel\"", got)
	}
	// 恰好等长不补零不截断
	buf.Reset()
	Bytes.WriteFixedString(&buf, 5, "hello")
	if buf.Len() != 5 {
		t.Errorf("等长应占 5 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "hello" {
		t.Errorf("等长往返 = %q, want \"hello\"", got)
	}
	// 空串全补零
	buf.Reset()
	Bytes.WriteFixedString(&buf, 4, "")
	if buf.Len() != 4 {
		t.Errorf("空串应占 4 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "" {
		t.Errorf("空串往返 = %q, want \"\"", got)
	}
}

/* ---------------- UTF-16LE 宽字符字符串 ---------------- */

func TestBytesUtf16String(t *testing.T) {
	var buf bytes.Buffer
	Bytes.WriteUtf16String(&buf, 4, "你好")
	// 长度 = 字符数 * 2
	if buf.Len() != 8 {
		t.Errorf("UTF-16 定长应占 size*2=8 字节, got %d", buf.Len())
	}
	// '你' = U+4F60，小端首两字节为 60 4F
	bs := buf.Bytes()
	if bs[0] != 0x60 || bs[1] != 0x4F {
		t.Errorf("UTF-16LE 字节序错误: % x, want 60 4f ...", bs)
	}
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "你好" {
		t.Errorf("Utf16String 往返 = %q, want \"你好\"", got)
	}
	// size 小于字符数时按字符截断
	buf.Reset()
	Bytes.WriteUtf16String(&buf, 1, "你好")
	if buf.Len() != 2 {
		t.Errorf("截断后应占 2 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "你" {
		t.Errorf("UTF-16 截断 = %q, want \"你\"", got)
	}
	// 不足补零（补 0 字符）
	buf.Reset()
	Bytes.WriteUtf16String(&buf, 4, "你")
	if buf.Len() != 8 {
		t.Errorf("补零后应占 8 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "你" {
		t.Errorf("补零往返 = %q, want \"你\"", got)
	}
	// 空串
	buf.Reset()
	Bytes.WriteUtf16String(&buf, 2, "")
	if buf.Len() != 4 {
		t.Errorf("空串应占 4 字节, got %d", buf.Len())
	}
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "" {
		t.Errorf("空串往返 = %q, want \"\"", got)
	}
}

/* ---------------- NumberToBytes 类型分派 ---------------- */

func TestBytesNumberToBytes(t *testing.T) {
	cases := []struct {
		val  interface{}
		size int
	}{
		{int8(-5), 1}, {uint8(200), 1},
		{int16(-2), 2}, {uint16(0xBEEF), 2},
		{int32(-3), 4}, {uint32(0xDEADBEEF), 4},
		{int64(-4), 8}, {uint64(0x0123456789ABCDEF), 8},
		{int(-5), 8}, {float32(1.5), 4}, {float64(-2.5), 8},
	}
	for _, c := range cases {
		got := Bytes.NumberToBytes(c.val, false)
		if got == nil {
			t.Errorf("NumberToBytes(%T %v) 不应返回 nil", c.val, c.val)
			continue
		}
		if len(got) != c.size {
			t.Errorf("NumberToBytes(%T %v) 长度 = %d, want %d", c.val, c.val, len(got), c.size)
		}
	}
	// 字节序分派正确
	if v := Bytes.NumberToBytes(int32(0x01020304), false); v[0] != 0x04 {
		t.Errorf("NumberToBytes 小端首字节 = %#x, want 0x04", v[0])
	}
	if v := Bytes.NumberToBytes(int32(0x01020304), true); v[0] != 0x01 {
		t.Errorf("NumberToBytes 大端首字节 = %#x, want 0x01", v[0])
	}
	// 与解码方法配合抽查数值
	if got := Bytes.BytesToInt32(Bytes.NumberToBytes(int32(-99), false), false); got != -99 {
		t.Errorf("int32 分派往返 = %d, want -99", got)
	}
	if got := Bytes.BytesToInt64(Bytes.NumberToBytes(int64(-7), true), true); got != -7 {
		t.Errorf("int64 大端分派往返 = %d, want -7", got)
	}
	if got := Bytes.BytesToFloat32(Bytes.NumberToBytes(float32(1.5), false), false); got != 1.5 {
		t.Errorf("float32 分派往返 = %v, want 1.5", got)
	}
	if got := Bytes.BytesToFloat64(Bytes.NumberToBytes(2.5, true), true); got != 2.5 {
		t.Errorf("float64 大端分派往返 = %v, want 2.5", got)
	}
	// 不支持类型返回 nil
	for _, v := range []interface{}{"str", true, nil, []int{1}} {
		if got := Bytes.NumberToBytes(v, false); got != nil {
			t.Errorf("NumberToBytes(%v) 不支持类型应返回 nil, got %v", v, got)
		}
	}
}
