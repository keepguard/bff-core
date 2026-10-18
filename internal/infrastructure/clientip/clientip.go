package clientip

import (
	"net"
	"net/http"
	"strings"
)

// Nota de Arquitetura (ADR 0001 - 18/10/2026):
// Headers de CDN (ex.: CF-Connecting-IP, True-Client-IP) foram removidos propositalmente
// porque o cluster (k3s) não possui CDN (Cloudflare/Akamai) na frente; todos os domínios
// resolvem diretamente para a VPS. Aceitar tais headers permitiria falsificação de IP por
// qualquer visitante. Se um dia entrar Cloudflare na frente do cluster, isso deve ser revisto
// em conjunto com o `trustedIPs` do Traefik e com a validação das faixas de IP da CDN.
//
// Na arquitetura atual:
// 1. O Traefik (borda com externalTrafficPolicy: Local) define o IP real da conexão no XFF.
// 2. O nginx do front repassa o último item do XFF para X-Real-IP e X-Forwarded-For.
// 3. Chamadores internos (BFFs de produto) passam o IP do visitante via X-Real-IP.
func FromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. X-Real-IP de chamador interno confiável vence
	if realIP := cleanIP(r.Header.Get("X-Real-IP")); realIP != "" {
		if parseIP(realIP) != nil {
			return realIP
		}
	}

	// 2. X-Forwarded-For: o IP confiável colocado pela borda (Traefik) é o ÚLTIMO item da cadeia
	if xff := r.Header.Get("X-Forwarded-For"); strings.TrimSpace(xff) != "" {
		if ip := lastPublic(xff); ip != "" {
			return ip
		}
		if ip := lastAny(xff); ip != "" {
			return ip
		}
	}

	// 3. Sem cabeçalhos válidos, vale o RemoteAddr da conexão
	if remote := cleanIP(r.RemoteAddr); remote != "" {
		if parseIP(remote) != nil {
			return remote
		}
	}

	return ""
}

func lastPublic(raw string) string {
	ips := splitIPs(raw)
	for i := len(ips) - 1; i >= 0; i-- {
		if parsed := parseIP(ips[i]); parsed != nil && !isPrivate(parsed) {
			return ips[i]
		}
	}
	return ""
}

func lastAny(raw string) string {
	ips := splitIPs(raw)
	for i := len(ips) - 1; i >= 0; i-- {
		if parseIP(ips[i]) != nil {
			return ips[i]
		}
	}
	return ""
}

func cleanIP(value string) string {
	ip := strings.TrimSpace(value)
	if ip == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return strings.Trim(ip, "[]")
}

func splitIPs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ips := make([]string, 0, len(parts))
	for _, part := range parts {
		ip := cleanIP(part)
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
