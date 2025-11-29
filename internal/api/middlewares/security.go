package middlewares

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// Strict Transport Security (HTTPS only)
		// Only set in production with HTTPS
		// c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy
		c.Writer.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' https://cdn.paddle.com https://sandbox-cdn.paddle.com; "+
				"style-src 'self' 'unsafe-inline' https://cdn.paddle.com https://sandbox-cdn.paddle.com; "+
				"img-src 'self' data: https:; "+
				"font-src 'self' data:; "+
				"connect-src 'self' https://cdn.paddle.com https://sandbox-cdn.paddle.com https://sandbox-api.paddle.com https://api.paddle.com https://sandbox-checkout.paddle.com https://checkout.paddle.com https://sandbox-buy.paddle.com https://buy.paddle.com; "+
				"frame-src https://sandbox-checkout.paddle.com https://checkout.paddle.com https://sandbox-buy.paddle.com https://buy.paddle.com; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")

		// Referrer Policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature Policy)
		c.Writer.Header().Set("Permissions-Policy",
			"accelerometer=(), "+
				"camera=(), "+
				"geolocation=(), "+
				"gyroscope=(), "+
				"magnetometer=(), "+
				"microphone=(), "+
				"payment=(self), "+
				"usb=()")

		// Cache control for sensitive endpoints
		if c.Request.URL.Path != "/api/health" && c.Request.URL.Path != "/metrics" {
			c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Writer.Header().Set("Pragma", "no-cache")
			c.Writer.Header().Set("Expires", "0")
		}

		c.Next()
	}
}
