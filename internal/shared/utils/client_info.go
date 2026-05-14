package utils

import "github.com/gofiber/fiber/v3"

type ClientInfo struct {
	UserAgent  string
	IPAddress  string
	ClientType string
}

func GetClientInfo(c fiber.Ctx) *ClientInfo {
	clientType := c.Get("X-Client-Type")

	if clientType == "" {
		clientType = "web"
	}

	return &ClientInfo{
		UserAgent:  c.Get("User-Agent"),
		IPAddress:  c.IP(),
		ClientType: clientType,
	}
}
