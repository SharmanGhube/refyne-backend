package middlewares

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing (CORS) for the API
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins based on environment
		allowedOrigins := getAllowedOrigins()

		if isOriginAllowed(origin, allowedOrigins) {
			// Always reflect the actual origin back — browsers reject credentials + wildcard "*"
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
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
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr != "" {
		origins := strings.Split(originsStr, ",")
		for i, o := range origins {
			origins[i] = strings.TrimSpace(o)
		}
		return origins
	}

	// Fallback to defaults
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3000",
		"https://refyne.app",
		"https://www.refyne.me",
		"https://app.refyne.me",
		"https://refyne-frontend.vercel.app",
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	// If allow-all wildcard is in allowedOrigins, allow any origin
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if origin == allowed {
			return true
		}
	}
	return false
}
