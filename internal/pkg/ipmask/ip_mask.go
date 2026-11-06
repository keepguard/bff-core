// Package ipmask trunca IP antes da auditoria (contrato C2 §13): IPv4 /24, IPv6 /48.
// Mesma regra do ms-billing-go.
package ipmask

import "net"

// Truncate devolve o IP truncado. O que não é IP sai vazio: nunca passa adiante texto
// que veio de cabeçalho.
func Truncate(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.Mask(net.CIDRMask(24, 32)).String()
	}
	return ip.Mask(net.CIDRMask(48, 128)).String()
}
