package middlewares

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing (CORS) for the API
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins based on environment
		allowedOrigins := getAllowedOrigins()

		if isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set allowed methods
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Set allowed headers
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, "+
				"X-Request-ID, X-Requested-With, Origin, Accept-Encoding, Accept-Language")

		// Set exposed headers (headers the frontend can read)
		c.Writer.Header().Set("Access-Control-Expose-Headers",
			"Content-Length, Content-Type, X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")

		// Set max age for preflight cache
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getAllowedOrigins returns the list of allowed origins based on environment
func getAllowedOrigins() []string {
	// In production, restrict to specific domains
	// In development, allow localhost
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3000",
		"https://refyne.app",
		"https://www.refyne.me",
		"https://app.refyne.me",
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}
