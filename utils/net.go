package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// utilNet 网络工具。
//
// 覆盖 IPv4 与整数互转、内网与 CIDR 判断、本机 IP/MAC 探测、端口可用性
// 与 TCP 探活等（对应 hutool NetUtil / Ipv4Util 的可移植子集）。
type utilNet struct{}

/* ---------------- IPv4 与整数互转 ---------------- */

// Ipv4ToLong 把点分十进制 IPv4 转为 uint64（对应 hutool ipv4ToLong，方法别名 inet_aton）。
// 要求恰好 4 段、每段为 0~255 的十进制数：Ipv4ToLong("192.168.1.1") = 3232235777，
// 非法格式返回 error。场景：IP 以整数落库、IP 区间数值化比较。
func (n *utilNet) Ipv4ToLong(ip string) (uint64, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("非法 IPv4 地址: %q", ip)
	}
	var result uint64
	for _, part := range parts {
		// 只接受纯数字（拒绝空段与 "+1"、"-1"、空白等符号前缀），取值 0~255
		if part == "" || part[0] < '0' || part[0] > '9' {
			return 0, fmt.Errorf("非法 IPv4 地址: %q", ip)
		}
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return 0, fmt.Errorf("非法 IPv4 地址: %q", ip)
		}
		result = result<<8 | uint64(num)
	}
	return result, nil
}

// LongToIpv4 把 uint64 还原为点分十进制 IPv4（对应 hutool longToIpv4，方法别名 inet_ntoa），只取低 32 位。
// LongToIpv4(3232235777) = "192.168.1.1"。场景：还原数据库中整数存储的 IP。
func (n *utilNet) LongToIpv4(ipLong uint64) string {
	return strconv.Itoa(int(ipLong>>24&0xFF)) + "." +
		strconv.Itoa(int(ipLong>>16&0xFF)) + "." +
		strconv.Itoa(int(ipLong>>8&0xFF)) + "." +
		strconv.Itoa(int(ipLong&0xFF))
}

/* ---------------- 内网与网段 ---------------- */

// 内网网段边界（hutool isInnerIP 的三段私有地址边界，按 32 位无符号换算）
const (
	netInnerABegin = 0x0A000000 // 10.0.0.0
	netInnerAEnd   = 0x0AFFFFFF // 10.255.255.255
	netInnerBBegin = 0xAC100000 // 172.16.0.0
	netInnerBEnd   = 0xAC1FFFFF // 172.31.255.255
	netInnerCBegin = 0xC0A80000 // 192.168.0.0
	netInnerCEnd   = 0xC0A8FFFF // 192.168.255.255
)

// IsInnerIP 判断是否为内网 IPv4（对应 hutool Ipv4Util.isInnerIP）：
//   - A 类 10.0.0.0 ~ 10.255.255.255
//   - B 类 172.16.0.0 ~ 172.31.255.255
//   - C 类 192.168.0.0 ~ 192.168.255.255
//   - 环回地址：hutool 仅精确匹配 "127.0.0.1"（127.0.0.2 不算内网），此处保持一致
//
// 非法 IP 返回 false。场景：区分内外网流量、后台接口内网白名单。
func (n *utilNet) IsInnerIP(ip string) bool {
	ipNum, err := n.Ipv4ToLong(ip)
	if err != nil {
		return false
	}
	return (ipNum >= netInnerABegin && ipNum <= netInnerAEnd) ||
		(ipNum >= netInnerBBegin && ipNum <= netInnerBEnd) ||
		(ipNum >= netInnerCBegin && ipNum <= netInnerCEnd) ||
		ip == "127.0.0.1"
}

// IsInRange 判断 IP 是否在 CIDR 网段内（对应 hutool NetUtil.isInRange，基于 net.ParseCIDR 实现）。
// IsInRange("192.168.1.10", "192.168.1.0/24") = true；ip 或 cidr 非法时返回 false。
// 场景：访问控制网段匹配、按 IP 段分流。
func (n *utilNet) IsInRange(ip, cidr string) bool {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return ipnet.Contains(parsed)
}

/* ---------------- 本机信息 ---------------- */

// LocalIps 返回本机所有非环回 IPv4 地址，多网卡按系统枚举顺序返回并去重；
// 枚举网卡失败返回 nil。场景：服务注册上报本机 IP、多网卡环境挑选绑定地址。
func (n *utilNet) LocalIps() []string {
	return netLocalIpv4s()
}

// GetLocalIP 返回本机 IPv4 地址，优选规则同 hutool getLocalhost：
// 先取第一个非内网（公网）地址，无公网地址时回退取第一个内网地址，均无返回空串。
// 场景：服务自注册、上报调用方来源地址。
func (n *utilNet) GetLocalIP() string {
	ips := netLocalIpv4s()
	for _, ip := range ips {
		if !n.IsInnerIP(ip) {
			return ip
		}
	}
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

// GetMacAddress 返回第一块非环回网卡的物理地址，冒号分隔小写十六进制，
// 如 "aa:bb:cc:dd:ee:ff"；获取失败（无可用网卡等）返回空串。
// 场景：机器指纹、设备绑定授权。
func (n *utilNet) GetMacAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || len(iface.HardwareAddr) == 0 {
			continue
		}
		// HardwareAddr.String() 即冒号分隔的小写十六进制形式
		return iface.HardwareAddr.String()
	}
	return ""
}

/* ---------------- 端口与探活 ---------------- */

// IsOpen TCP 探活：向 hostPort（形如 "127.0.0.1:8080"、"example.com:443"）发起 TCP 连接，
// 连接成功视为端口开放，随即关闭连接。timeout<=0 时使用默认 2 秒。
// 说明：hutool isOpen 即以 Socket 连接探测远程端口；Go 标准库不便发 ICMP，
// 故此处同样以 TCP 连接探测替代 ping。场景：服务存活检查、发布前端口连通性验证。
func (n *utilNet) IsOpen(hostPort string, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	conn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// IsValidPort 判断端口号是否有效（1~65535）。
// 注意：hutool 同名方法允许 0（交由系统分配临时端口），此处按常用约定从 1 起算；
// 仅校验范围，不检查占用（占用检测见 UsableLocalPort）。
func (n *utilNet) IsValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

// UsableLocalPort 从 minPort 起顺序查找第一个可用端口（上限 65535）。
// minPort 越界或范围内端口全部被占用时返回 error。
// 场景：内嵌 HTTP/调试服务启动前动态挑选空闲端口。
func (n *utilNet) UsableLocalPort(minPort int) (int, error) {
	ports, err := n.UsableLocalPorts(1, minPort, 65535)
	if err != nil {
		return 0, err
	}
	return ports[0], nil
}

// UsableLocalPorts 在 [minPort, maxPort] 内顺序查找 num 个可用端口，结果按端口号升序返回。
// num<=0、范围非法（minPort<1、maxPort>65535、minPort>maxPort）或可用端口不足 num 个时返回 error。
// 场景：同时拉起多个本地服务（业务口 + 管理口）时批量预留端口。
func (n *utilNet) UsableLocalPorts(num, minPort, maxPort int) ([]int, error) {
	if num <= 0 {
		return nil, fmt.Errorf("需要查找的端口数量必须大于 0: %d", num)
	}
	if minPort < 1 || maxPort > 65535 || minPort > maxPort {
		return nil, fmt.Errorf("非法端口范围: [%d, %d]", minPort, maxPort)
	}
	result := make([]int, 0, num)
	for port := minPort; port <= maxPort && len(result) < num; port++ {
		if netUsablePort(port) {
			result = append(result, port)
		}
	}
	if len(result) < num {
		return nil, fmt.Errorf("范围 [%d, %d] 内仅找到 %d 个可用端口（需要 %d 个）", minPort, maxPort, len(result), num)
	}
	return result, nil
}

/* ---------------- 代理头与 Cookie ---------------- */

// HideIpPart 将 IP 末段打码为 *（对应 hutool hideIpPart）。
// HideIpPart("202.112.14.33") = "202.112.14.*"；不含 "." 的输入返回 "*"，空串返回空串。
// 场景：日志、公告中对用户 IP 脱敏。
func (n *utilNet) HideIpPart(ip string) string {
	if ip == "" {
		return ""
	}
	// LastIndex 为 -1 时 ip[:0] + "*" = "*"，与 hutool 行为一致
	return ip[:strings.LastIndex(ip, ".")+1] + "*"
}

// GetIpByHost 解析域名并返回第一个 IPv4 地址，GetIpByHost("localhost") = "127.0.0.1"。
// 解析失败或结果中无 IPv4 记录时返回空串
// （注意：hutool 同名方法失败时原样返回域名，此处按"失败给空值"约定返回空串）。
// 场景：动态域名取 IP、内网服务寻址。
func (n *utilNet) GetIpByHost(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

// GetMultistageReverseProxyIp 从多级反向代理头值（形如 "10.8.0.1, 210.22.22.153"）中
// 取第一个非 unknown 的项作为客户端 IP（对应 hutool getMultistageReverseProxyIp）。
// 不含逗号时原样返回。场景：解析 X-Forwarded-For 等代理链还原真实来源。
func (n *utilNet) GetMultistageReverseProxyIp(ip string) string {
	if strings.Index(ip, ",") > 0 {
		for _, subIp := range strings.Split(ip, ",") {
			if subIp = strings.TrimSpace(subIp); !netIsUnknown(subIp) {
				return subIp
			}
		}
	}
	return ip
}

// ParseCookies 解析 Cookie 请求头（形如 "a=1; b=2"）为键值对 map。
// 以 ";" 分段、按第一个 "=" 切分键值，无 "=" 或键为空的段忽略，重复键后者覆盖；
// 空白输入返回空 map。场景：服务端解析请求 Cookie、爬虫维持会话。
func (n *utilNet) ParseCookies(headerValue string) map[string]string {
	result := make(map[string]string)
	for _, seg := range strings.Split(headerValue, ";") {
		seg = strings.TrimSpace(seg)
		idx := strings.Index(seg, "=")
		if seg == "" || idx <= 0 {
			// 跳过空段、无 "=" 的段与空键段（对应 hutool HttpCookie.parse 跳过非法项）
			continue
		}
		result[seg[:idx]] = seg[idx+1:]
	}
	return result
}

/* ---------------- 内部辅助（net 前缀，避免与其它文件冲突） ---------------- */

// netLocalIpv4s 遍历所有网卡，按系统枚举顺序返回非环回接口上的 IPv4 地址
// （去重，保留首次出现顺序）；枚举失败返回 nil
func netLocalIpv4s() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	ips := make([]string, 0, len(interfaces))
	seen := make(map[string]struct{}, len(interfaces))
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			ipv4 := ip.To4()
			if ipv4 == nil || ipv4.IsLoopback() {
				continue
			}
			key := ipv4.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ips = append(ips, key)
		}
	}
	return ips
}

// netUsablePort 检测本地端口当前是否可绑定：TCP 与 UDP 都能绑定才算可用
// （对应 hutool isUsableLocalPort 同时探测 ServerSocket 与 DatagramSocket）
func netUsablePort(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	tcp, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = tcp.Close()
	udp, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = udp.Close()
	return true
}

// netIsUnknown 判断代理头字段是否为"未知"：空白串或不区分大小写的 "unknown"
func netIsUnknown(checkString string) bool {
	return strings.TrimSpace(checkString) == "" || strings.EqualFold(checkString, "unknown")
}
