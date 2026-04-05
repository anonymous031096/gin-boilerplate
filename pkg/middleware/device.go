package middleware

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const (
	HeaderDeviceID  = "X-Device-Id"
	DefaultDeviceID = "unknown"
)

// GetDeviceFingerprint generates SHA-256 hash from deviceID + IP + User-Agent.
// Same client always produces same hash. Different IP or browser → different hash.
func GetDeviceFingerprint(c *gin.Context) string {
	deviceID := c.GetHeader(HeaderDeviceID)
	if deviceID == "" {
		deviceID = DefaultDeviceID
	}

	raw := deviceID + "|" + c.ClientIP() + "|" + c.GetHeader("User-Agent")
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
