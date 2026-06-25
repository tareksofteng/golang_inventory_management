package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS allows the Vue dev server (a different origin/port) to call this API from
// the browser. For a demo we allow any origin; in production you would restrict
// Access-Control-Allow-Origin to your real frontend domain.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		// Browsers send a preflight OPTIONS request before the real one.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
