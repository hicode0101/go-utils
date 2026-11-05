package utils

import (
	"fmt"
	"net"
	"regexp"
	"testing"
	"time"
)

/* ---------------- IPv4 与整数互转（hutool Ipv4Util） ---------------- */

func TestNetIpv4ToLong(t *testing.T) {
	// 黄金值（同 hutool ipv4ToLong 测试）
	if got, err := Net.Ipv4ToLong("192.168.1.1"); err != nil || got != 3232235777 {
		t.Errorf("Ipv4ToLong(\"192.168.1.1\") = %d, %v, want 3232235777", got, err)
	}
	if got, err := Net.Ipv4ToLong("0.0.0.0"); err != nil || got != 0 {
		t.Errorf("Ipv4ToLong(\"0.0.0.0\") = %d, %v, want 0", got, err)
	}
	if got, err := Net.Ipv4ToLong("255.255.255.255"); err != nil || got != 4294967295 {
		t.Errorf("Ipv4ToLong(\"255.255.255.255\") = %d, %v, want 4294967295", got, err)
	}
	if got, err := Net.Ipv4ToLong("10.0.0.138"); err != nil || got != 167772298 {
		t.Errorf("Ipv4ToLong(\"10.0.0.138\") = %d, %v, want 167772298", got, err)
	}

	// 非法格式：段数不对、越界、空段、负号/符号前缀、非数字
	for _, bad := range []string{
		"1.2.3", "1.2.3.4.5", "1.2.3.256", "256.1.1.1",
		"1.2.3.", ".1.2.3", "", "a.b.c.d", " 1.2.3.4", "+1.2.3.4", "-1.2.3.4",
	} {
		if got, err := Net.Ipv4ToLong(bad); err == nil {
			t.Errorf("Ipv4ToLong(%q) = %d, 应返回错误", bad, got)
		}
	}
}

func TestNetLongToIpv4(t *testing.T) {
	// 黄金值与边界
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0.0.0.0"},
		{4294967295, "255.255.255.255"},
		{3232235777, "192.168.1.1"},
		{167772298, "10.0.0.138"},
		// 只取低 32 位：高位应被忽略
		{1 << 32, "0.0.0.0"},
		{1<<32 | 4294967295, "255.255.255.255"},
	}
	for _, c := range cases {
		if got := Net.LongToIpv4(c.in); got != c.want {
			t.Errorf("LongToIpv4(%d) = %q, want %q", c.in, got, c.want)
		}
	}

	// 往返：LongToIpv4(Ipv4ToLong(x)) == x
	for _, ip := range []string{"192.168.1.1", "0.0.0.0", "255.255.255.255", "10.0.0.138", "172.31.255.255", "1.2.3.4"} {
		long, err := Net.Ipv4ToLong(ip)
		if err != nil {
			t.Fatalf("Ipv4ToLong(%q) 意外失败: %v", ip, err)
		}
		if got := Net.LongToIpv4(long); got != ip {
			t.Errorf("往返 %q: LongToIpv4(%d) = %q", ip, long, got)
		}
	}
}

/* ---------------- 内网与网段 ---------------- */

func TestNetIsInnerIP(t *testing.T) {
	// 内网：A/B/C 类私有段 + 环回
	for _, ip := range []string{
		"10.0.0.0", "10.1.2.3", "10.255.255.255",
		"172.16.0.0", "172.20.1.2", "172.31.255.255",
		"192.168.0.0", "192.168.1.100", "192.168.255.255",
		"127.0.0.1",
	} {
		if !Net.IsInnerIP(ip) {
			t.Errorf("IsInnerIP(%q) = false, want true", ip)
		}
	}
	// 外网与边界外：172.15/172.32 不在 B 类段；hutool 语义下仅精确匹配 "127.0.0.1"
	for _, ip := range []string{
		"8.8.8.8", "11.0.0.1", "172.15.255.255", "172.32.0.1", "192.169.0.1", "127.0.0.2",
	} {
		if Net.IsInnerIP(ip) {
			t.Errorf("IsInnerIP(%q) = true, want false", ip)
		}
	}
	// 非法 IP 返回 false
	for _, ip := range []string{"not-an-ip", "256.1.1.1", "1.2.3", ""} {
		if Net.IsInnerIP(ip) {
			t.Errorf("IsInnerIP(%q) = true, 非法输入应为 false", ip)
		}
	}
}

func TestNetIsInRange(t *testing.T) {
	// 命中：普通地址、网络地址本身、广播地址
	for _, ip := range []string{"192.168.1.10", "192.168.1.0", "192.168.1.255"} {
		if !Net.IsInRange(ip, "192.168.1.0/24") {
			t.Errorf("IsInRange(%q, \"192.168.1.0/24\") = false, want true", ip)
		}
	}
	// 不命中
	for _, ip := range []string{"10.0.0.1", "192.168.2.10"} {
		if Net.IsInRange(ip, "192.168.1.0/24") {
			t.Errorf("IsInRange(%q, \"192.168.1.0/24\") = true, want false", ip)
		}
	}
	// 非法 cidr 或非法 ip 均返回 false
	if Net.IsInRange("192.168.1.10", "192.168.1.0/33") {
		t.Error("IsInRange 非法 cidr（/33）应返回 false")
	}
	if Net.IsInRange("192.168.1.10", "not-a-cidr") {
		t.Error("IsInRange 非法 cidr 应返回 false")
	}
	if Net.IsInRange("not-an-ip", "192.168.1.0/24") {
		t.Error("IsInRange 非法 ip 应返回 false")
	}
}

/* ---------------- 本机信息（基于本机环境，不访问外网） ---------------- */

func TestNetLocalIps(t *testing.T) {
	ips := Net.LocalIps()
	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed == nil || parsed.To4() == nil || parsed.IsLoopback() {
			t.Errorf("LocalIps 含非法或环回地址: %q", ip)
		}
	}

	// GetLocalIP：为空时环境相关不判失败；非空必须是合法 IPv4
	lip := Net.GetLocalIP()
	if lip == "" {
		t.Log("本机无可用 IPv4（环境相关，不判失败）")
		return
	}
	parsed := net.ParseIP(lip)
	if parsed == nil || parsed.To4() == nil {
		t.Errorf("GetLocalIP = %q, 不是合法 IPv4", lip)
	}
}

func TestNetGetMacAddress(t *testing.T) {
	mac := Net.GetMacAddress()
	if mac == "" {
		t.Log("本机无可用非环回网卡 MAC（环境相关，不判失败）")
		return
	}
	// 冒号分隔的小写十六进制，如 "aa:bb:cc:dd:ee:ff"
	ok, err := regexp.MatchString(`^([0-9a-f]{2}):([0-9a-f]{2}):([0-9a-f]{2}):([0-9a-f]{2}):([0-9a-f]{2}):([0-9a-f]{2})$`, mac)
	if err != nil || !ok {
		t.Errorf("GetMacAddress = %q, 不符合 xx:xx:xx:xx:xx:xx 小写十六进制格式", mac)
	}
}

/* ---------------- 端口与探活（全部基于 127.0.0.1 回环） ---------------- */

func TestNetIsOpen(t *testing.T) {
	// 本地起监听 → 探活应成功
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("本地监听失败: %v", err)
	}
	addr := l.Addr().String()
	if !Net.IsOpen(addr, time.Second) {
		t.Errorf("IsOpen(%q, 1s) = false, 本地监听端口应开放", addr)
	}
	// timeout<=0 走默认 2 秒路径
	if !Net.IsOpen(addr, 0) {
		t.Errorf("IsOpen(%q, 0) = false, timeout<=0 应使用默认超时并探活成功", addr)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("关闭监听失败: %v", err)
	}

	// 换一个确认空闲的端口（监听拿号后立即释放）→ 探活应失败；负超时同样走默认路径
	f, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("获取空闲端口失败: %v", err)
	}
	freePort := f.Addr().(*net.TCPAddr).Port
	if err := f.Close(); err != nil {
		t.Fatalf("释放空闲端口失败: %v", err)
	}
	freeAddr := fmt.Sprintf("127.0.0.1:%d", freePort)
	if Net.IsOpen(freeAddr, time.Second) {
		t.Errorf("IsOpen(%q, 1s) = true, 确认空闲的端口不应开放", freeAddr)
	}
	if Net.IsOpen(freeAddr, -time.Second) {
		t.Errorf("IsOpen(%q, -1s) = true, 负超时应走默认路径并返回 false", freeAddr)
	}
}

func TestNetIsValidPort(t *testing.T) {
	if Net.IsValidPort(0) {
		t.Error("IsValidPort(0) = true, want false（按约定从 1 起算）")
	}
	if !Net.IsValidPort(1) {
		t.Error("IsValidPort(1) = false, want true")
	}
	if !Net.IsValidPort(65535) {
		t.Error("IsValidPort(65535) = false, want true")
	}
	if Net.IsValidPort(65536) {
		t.Error("IsValidPort(65536) = true, want false")
	}
	if Net.IsValidPort(-1) {
		t.Error("IsValidPort(-1) = true, want false")
	}
	if !Net.IsValidPort(8080) {
		t.Error("IsValidPort(8080) = false, want true")
	}
}

func TestNetUsableLocalPort(t *testing.T) {
	// 通配绑定占住一个随机空闲端口，UsableLocalPort 必须跳过它返回更大的
	w, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("通配监听失败: %v", err)
	}
	occupied := w.Addr().(*net.TCPAddr).Port
	defer w.Close() //nolint:errcheck

	got, err := Net.UsableLocalPort(occupied)
	if err != nil {
		t.Fatalf("UsableLocalPort(%d) 意外失败: %v", occupied, err)
	}
	if got <= occupied || got > 65535 {
		t.Errorf("UsableLocalPort(%d) = %d, 应跳过被占用端口返回更大值", occupied, got)
	}

	// 批量：范围 [occupied, 65535] 内取 2 个，互不相同且升序、都大于被占用端口
	ports, err := Net.UsableLocalPorts(2, occupied, 65535)
	if err != nil {
		t.Fatalf("UsableLocalPorts(2, %d, 65535) 意外失败: %v", occupied, err)
	}
	if len(ports) != 2 || ports[0] == ports[1] || ports[0] <= occupied || ports[1] <= ports[0] {
		t.Errorf("UsableLocalPorts(2, %d, 65535) = %v, 应为互不相同的升序端口且都大于 %d", occupied, ports, occupied)
	}

	// 非法参数
	if _, err := Net.UsableLocalPort(0); err == nil {
		t.Error("UsableLocalPort(0) 应返回错误（minPort 越界）")
	}
	if _, err := Net.UsableLocalPort(65536); err == nil {
		t.Error("UsableLocalPort(65536) 应返回错误（minPort 越界）")
	}
	if _, err := Net.UsableLocalPorts(0, 1, 65535); err == nil {
		t.Error("UsableLocalPorts(num=0) 应返回错误")
	}
	if _, err := Net.UsableLocalPorts(-1, 1, 65535); err == nil {
		t.Error("UsableLocalPorts(num=-1) 应返回错误")
	}
	if _, err := Net.UsableLocalPorts(1, 0, 65535); err == nil {
		t.Error("UsableLocalPorts(minPort=0) 应返回错误")
	}
	if _, err := Net.UsableLocalPorts(1, 1, 65536); err == nil {
		t.Error("UsableLocalPorts(maxPort=65536) 应返回错误")
	}
	if _, err := Net.UsableLocalPorts(2, 200, 100); err == nil {
		t.Error("UsableLocalPorts(minPort>maxPort) 应返回错误")
	}
}

/* ---------------- 代理头与 Cookie ---------------- */

func TestNetHideIpPart(t *testing.T) {
	if got := Net.HideIpPart("192.168.1.100"); got != "192.168.1.*" {
		t.Errorf("HideIpPart(\"192.168.1.100\") = %q, want \"192.168.1.*\"", got)
	}
	if got := Net.HideIpPart("202.112.14.33"); got != "202.112.14.*" {
		t.Errorf("HideIpPart(\"202.112.14.33\") = %q, want \"202.112.14.*\"", got)
	}
	if got := Net.HideIpPart("nodot"); got != "*" {
		t.Errorf("HideIpPart(\"nodot\") = %q, want \"*\"", got)
	}
	if got := Net.HideIpPart(""); got != "" {
		t.Errorf("HideIpPart(\"\") = %q, want 空串", got)
	}
}

func TestNetGetIpByHost(t *testing.T) {
	// 本机回环语义：字面量 IP 解析必返回自身
	if got := Net.GetIpByHost("127.0.0.1"); got == "" {
		t.Error("GetIpByHost(\"127.0.0.1\") = 空串, want 非空")
	} else if got != "127.0.0.1" {
		t.Logf("GetIpByHost(\"127.0.0.1\") = %q（环境相关，不判失败）", got)
	}
	// localhost 走本机 hosts 解析，不发外网请求
	if got := Net.GetIpByHost("localhost"); got == "" {
		t.Error("GetIpByHost(\"localhost\") = 空串, want 非空")
	} else if got != "127.0.0.1" {
		t.Logf("GetIpByHost(\"localhost\") = %q（环境相关，不判失败）", got)
	}
	// 保留 TLD ".invalid" 保证不解析 → 空串
	if got := Net.GetIpByHost("this-host-does-not-exist-xyz.invalid"); got != "" {
		t.Errorf("GetIpByHost(不存在的域名) = %q, want 空串", got)
	}
}

func TestNetGetMultistageReverseProxyIp(t *testing.T) {
	// 多层：取第一个非 unknown 项
	if got := Net.GetMultistageReverseProxyIp("unknown, 210.22.22.153, 10.8.0.1"); got != "210.22.22.153" {
		t.Errorf("GetMultistageReverseProxyIp = %q, want \"210.22.22.153\"", got)
	}
	if got := Net.GetMultistageReverseProxyIp("10.8.0.1, unknown"); got != "10.8.0.1" {
		t.Errorf("GetMultistageReverseProxyIp = %q, want \"10.8.0.1\"", got)
	}
	// unknown 大小写不敏感
	if got := Net.GetMultistageReverseProxyIp("UNKNOWN, 10.8.0.1"); got != "10.8.0.1" {
		t.Errorf("GetMultistageReverseProxyIp = %q, want \"10.8.0.1\"", got)
	}
	// 全 unknown：原样返回
	if got := Net.GetMultistageReverseProxyIp("unknown, UNKNOWN"); got != "unknown, UNKNOWN" {
		t.Errorf("GetMultistageReverseProxyIp(全 unknown) = %q, want 原样返回", got)
	}
	// 单 IP（无逗号）：原样返回
	if got := Net.GetMultistageReverseProxyIp("192.168.1.5"); got != "192.168.1.5" {
		t.Errorf("GetMultistageReverseProxyIp(单 IP) = %q, want \"192.168.1.5\"", got)
	}
	if got := Net.GetMultistageReverseProxyIp(""); got != "" {
		t.Errorf("GetMultistageReverseProxyIp(\"\") = %q, want 空串", got)
	}
}

func TestNetParseCookies(t *testing.T) {
	m := Net.ParseCookies("a=1; b=2")
	if len(m) != 2 || m["a"] != "1" || m["b"] != "2" {
		t.Errorf("ParseCookies(\"a=1; b=2\") = %v, want map[a:1 b:2]", m)
	}
	// 空串 → 空 map（非 nil）
	empty := Net.ParseCookies("")
	if empty == nil || len(empty) != 0 {
		t.Errorf("ParseCookies(\"\") = %v, want 非nil空map", empty)
	}
	// 非法段跳过：无 "="、"=值"（空键）、空段；重复键后者覆盖；值可含 "="
	m2 := Net.ParseCookies("flag; =novalue; k=v; a=1; a=2; token=abc=def; ")
	if len(m2) != 3 || m2["k"] != "v" || m2["a"] != "2" || m2["token"] != "abc=def" {
		t.Errorf("ParseCookies 非法段/覆盖/含等号值 = %v", m2)
	}
}
