package middleware

import "github.com/gin-gonic/gin"

const (
	HeaderDeviceID  = "X-Device-Id"
	DefaultDeviceID = "unknown"
)

func GetDeviceID(c *gin.Context) string {
	deviceID := c.GetHeader(HeaderDeviceID)
	if deviceID == "" {
		return DefaultDeviceID
	}
	return deviceID
}
