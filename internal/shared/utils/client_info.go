package utils

import (
	"strings"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	"github.com/gofiber/fiber/v3"
)

type ClientInfo struct {
	UserAgent  string
	IPAddress  string
	ClientType string
}

func GetClientInfo(c fiber.Ctx) *ClientInfo {
	clientType := strings.ToLower(c.Get("X-Client-Type"))

	switch clientType {
	case constants.ClientTypeMobile:
		clientType = constants.ClientTypeMobile
	default:
		clientType = constants.ClientTypeWeb
	}

	return &ClientInfo{
		UserAgent:  c.Get("User-Agent"),
		IPAddress:  getRealIP(c),
		ClientType: clientType,
	}
}

func getRealIP(c fiber.Ctx) string {
	// Cloudflare
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Standard Reverse Proxy (Nginx, AWS ALB, etc.)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		if strings.Contains(xff, ",") {
			parts := strings.Split(xff, ",")
			return strings.TrimSpace(parts[0])
		}
		return xff
	}

	// Another proxy alternative
	if rip := c.Get("X-Real-IP"); rip != "" {
		return rip
	}

	return c.IP()
}
