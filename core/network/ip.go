package network

import (
	"github.com/pkg/errors"
	"go-com/core/logr"
	"math"
	"net"
	"strconv"
	"strings"
)

// IPv4RuleToCIDR 将IPv4规则转换为CIDR
func IPv4RuleToCIDR(rule string) (*net.IPNet, error) {
	var err error
	var ipNet *net.IPNet
	if strings.Contains(rule, "/") {
		_, ipNet, err = net.ParseCIDR(rule)
		if err != nil {
			return nil, errors.New("IP规则有误：" + rule + "。" + err.Error())
		} else {
			return ipNet, nil
		}
	}

	parts := strings.Split(rule, ".")
	if len(parts) != 4 {
		return nil, errors.New("IP规则有误：" + rule + "。地址段长度错误。")
	}

	var cidr string
	var num int
	mask := 32 // 默认掩码长度
	for i, part := range parts {
		switch part {
		case "*":
			num = 0
			mask -= 8 // 每段占用8位
		default:
			num, err = IPv4ParseSegment(part)
			if err != nil {
				return nil, err
			}
		}
		cidr += strconv.Itoa(num)
		if i != 3 {
			cidr += "."
		}
	}

	cidr += "/" + strconv.Itoa(mask)
	_, ipNet, err = net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.New("IP规则有误：" + rule + "。" + cidr)
	} else {
		return ipNet, nil
	}
}

// IPv4ParseSegment 解析IPv4段，确保其在0-255范围内
func IPv4ParseSegment(segment string) (int, error) {
	num, err := strconv.Atoi(segment)
	if err != nil || num < 0 || num > 255 {
		return 0, errors.New("无效的IP段：" + segment)
	}
	return num, nil
}

// IPv4CheckWhitelist 验证IPv4是否在白名单中
func IPv4CheckWhitelist(ip string, rules []string) bool {
	var err error
	var ipNet *net.IPNet
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, rule := range rules {
		if rule == "*" {
			return true
		}
		ipNet, err = IPv4RuleToCIDR(rule)
		if err != nil {
			logr.L.Error(err)
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// ParseAddr 解析ip地址
func ParseAddr(addr string) (string, int, error) {
	addrPart := strings.Split(addr, ":")
	addrPartLen := len(addrPart)
	if addrPartLen < 2 {
		return "", 0, errors.New("addr有误。")
	}
	host := strings.Join(addrPart[0:addrPartLen-1], ":")
	port := addrPart[addrPartLen-1]
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, errors.New("addr的端口有误。")
	}
	return host, portNum, nil
}

// IPString2Long 把ip字符串转为数值
func IPString2Long(ip string) (uint, error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
}

// Long2IPString 把数值转为ip字符串
func Long2IPString(i uint) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip.String(), nil
}
