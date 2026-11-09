package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if path == "/admin" || strings.HasPrefix(path, "/admin/") || strings.HasPrefix(path, "/api/") {
			c.Header("Cache-Control", "no-store")
		} else {
			c.Header("Cache-Control", "private, no-cache")
		}

		c.Next()
	}
}
