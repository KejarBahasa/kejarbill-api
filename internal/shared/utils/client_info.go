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
		IPAddress:  c.IP(),
		ClientType: clientType,
	}
}
