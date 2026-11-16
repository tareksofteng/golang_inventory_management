package middleware

import (
	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"github.com/gin-gonic/gin"
)

// mutatingMethods are the HTTP verbs that change data and are therefore worth
// recording in the audit trail. Reads (GET/HEAD/OPTIONS) are deliberately not
// logged — they do not alter state and would drown the trail in noise.
var mutatingMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

// AuditLog records every data-changing request after it completes. It runs the
// handler first (c.Next), so by the time it writes the entry the Auth middleware
// has already put the user id in context and the final status is known. Writing
// the entry is best-effort: a failure here must never break the user's request.
func AuditLog(repo repositories.AuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !mutatingMethods[c.Request.Method] {
			return
		}

		// Only authenticated requests are audited — the trail answers "which user
		// changed what", and every data mutation requires a logged-in user. This
		// also keeps the public auth endpoints out of the trail.
		userID := UserID(c)
		if userID == 0 {
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // unmatched route (e.g. 404)
		}

		_ = repo.Create(&models.AuditLog{
			UserID: userID,
			Method: c.Request.Method,
			Path:   path,
			Status: c.Writer.Status(),
			IP:     c.ClientIP(),
		})
	}
}
