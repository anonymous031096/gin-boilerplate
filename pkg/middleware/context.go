package middleware

import "github.com/gin-gonic/gin"

const (
	keyUserID      = "user_id"
	keyDeviceID    = "device_id"
	keyPermissions = "permissions"
)

func SetCurrentUser(c *gin.Context, userID string, deviceID string, permissions []string) {
	c.Set(keyUserID, userID)
	c.Set(keyDeviceID, deviceID)
	c.Set(keyPermissions, permissions)
}

func GetCurrentUserID(c *gin.Context) string {
	return c.GetString(keyUserID)
}

func GetCurrentDeviceID(c *gin.Context) string {
	return c.GetString(keyDeviceID)
}

func GetCurrentPermissions(c *gin.Context) []string {
	perms, _ := c.Get(keyPermissions)
	if p, ok := perms.([]string); ok {
		return p
	}
	return nil
}
