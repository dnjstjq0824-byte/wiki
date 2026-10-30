package auth

import (
	"net"
	"strings"
)

// MaskIP scrubs the sensitive client host portions of IPv4 and IPv6 addresses
func MaskIP(ipStr string) string {
	cleanIP := strings.TrimSpace(ipStr)
	// Strip port if present
	if host, _, err := net.SplitHostPort(cleanIP); err == nil {
		cleanIP = host
	}

	parsed := net.ParseIP(cleanIP)
	if parsed == nil {
		return "xxx.xxx.xxx.xxx"
	}

	// Check if IPv4
	if ipv4 := parsed.To4(); ipv4 != nil {
		parts := strings.Split(ipv4.String(), ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + "." + parts[2] + ".xxx"
		}
		return "xxx.xxx.xxx.xxx"
	}

	// IPv6: Retain only first 64 bits (routing prefix)
	mask := net.CIDRMask(64, 128)
	network := parsed.Mask(mask)
	return network.String() + "/64"
}

