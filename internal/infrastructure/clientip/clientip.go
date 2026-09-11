package clientip

import (
	"net"
	"net/http"
	"strings"
)

var headerCandidates = []string{
	"CF-Connecting-IP",
	"True-Client-IP",
	"X-Real-IP",
	"X-Forwarded-For",
}

// FromRequest extrai o endereço IP seguro do cliente ignorando headers arbitrários do cliente (RN-RL-01).
func FromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	var fallback string
	for _, name := range headerCandidates {
		raw := r.Header.Get(name)
		if ip := firstPublic(raw); ip != "" {
			return ip
		}
		if fallback == "" {
			fallback = firstAny(raw)
		}
	}

	if ip := firstPublic(r.RemoteAddr); ip != "" {
		return ip
	}
	if fallback != "" {
		return fallback
	}
	return firstAny(r.RemoteAddr)
}

func firstPublic(raw string) string {
	for _, ip := range splitIPs(raw) {
		if parsed := parseIP(ip); parsed != nil && !isPrivate(parsed) {
			return ip
		}
	}
	return ""
}

func firstAny(raw string) string {
	for _, ip := range splitIPs(raw) {
		if parseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

func splitIPs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ips := make([]string, 0, len(parts))
	for _, part := range parts {
		ip := strings.TrimSpace(part)
		if ip == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		ip = strings.Trim(ip, "[]")
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func parseIP(value string) net.IP {
	return net.ParseIP(value)
}

func isPrivate(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified()
}
