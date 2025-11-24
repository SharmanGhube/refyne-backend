package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

// DeviceInfo contains parsed device information from request
type DeviceInfo struct {
	Fingerprint string
	DeviceName  string
	DeviceType  string // mobile, tablet, desktop, unknown
	Browser     string
	OS          string
	IPAddress   string
	UserAgent   string
}

// GenerateDeviceFingerprint creates a unique identifier for a device
// based on User-Agent and IP address
func GenerateDeviceFingerprint(userAgent, ipAddress string) string {
	// Normalize inputs
	ua := strings.TrimSpace(strings.ToLower(userAgent))
	ip := strings.TrimSpace(ipAddress)

	// Create hash of user-agent + IP
	data := fmt.Sprintf("%s|%s", ua, ip)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ExtractDeviceInfo parses request headers to extract device information
func ExtractDeviceInfo(c *gin.Context) *DeviceInfo {
	userAgent := c.GetHeader("User-Agent")
	ipAddress := GetClientIP(c)

	// Parse user agent
	ua := useragent.Parse(userAgent)

	// Determine device type
	deviceType := "desktop"
	if ua.Mobile {
		deviceType = "mobile"
	} else if ua.Tablet {
		deviceType = "tablet"
	} else if ua.Desktop {
		deviceType = "desktop"
	} else {
		deviceType = "unknown"
	}

	// Generate device name
	deviceName := fmt.Sprintf("%s on %s", ua.Name, ua.OS)
	if deviceName == " on " {
		deviceName = "Unknown Device"
	}

	return &DeviceInfo{
		Fingerprint: GenerateDeviceFingerprint(userAgent, ipAddress),
		DeviceName:  deviceName,
		DeviceType:  deviceType,
		Browser:     ua.Name,
		OS:          ua.OS,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}
}

// GetClientIP extracts the real client IP from request headers
// Handles X-Forwarded-For, X-Real-IP headers
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && isValidIP(xRealIP) {
		return xRealIP
	}

	// Fall back to RemoteAddr
	ip := c.ClientIP()
	if isValidIP(ip) {
		return ip
	}

	return c.Request.RemoteAddr
}

// isValidIP checks if the string is a valid IP address
func isValidIP(ip string) bool {
	// Remove port if present
	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)
		if err == nil {
			ip = host
		}
	}
	return net.ParseIP(ip) != nil
}

// IsSuspiciousLogin checks if a login attempt is suspicious based on various factors
func IsSuspiciousLogin(newIP string, knownIPs []string, newDeviceFingerprint string, knownFingerprints []string) (bool, string) {
	// Check if IP is known
	ipKnown := false
	for _, knownIP := range knownIPs {
		if newIP == knownIP {
			ipKnown = true
			break
		}
	}

	// Check if device is known
	deviceKnown := false
	for _, knownFP := range knownFingerprints {
		if newDeviceFingerprint == knownFP {
			deviceKnown = true
			break
		}
	}

	// Suspicious if both IP and device are new
	if !ipKnown && !deviceKnown {
		return true, "Login from new location and new device"
	}

	// Suspicious if only device is new but from known IP (possible device theft)
	if ipKnown && !deviceKnown {
		return true, "Login from new device at known location"
	}

	// Not suspicious if device is known (even from new IP - could be traveling)
	return false, ""
}

// CalculateIPDistance calculates approximate distance between two locations
// This is a simplified version - in production, use a proper geolocation service
func CalculateIPDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula for great-circle distance
	const earthRadius = 6371 // kilometers

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	a := 0.5 - 0.5*cosine(dLat) + cosine(degreesToRadians(lat1))*cosine(degreesToRadians(lat2))*(0.5-0.5*cosine(dLon))

	return 2 * earthRadius * asin(a)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * 3.14159265359 / 180
}

func cosine(x float64) float64 {
	// Simple approximation
	return 1 - (x*x)/2 + (x*x*x*x)/24
}

func asin(x float64) float64 {
	// Simple approximation for small values
	if x < -1 || x > 1 {
		return 0
	}
	return x + (x*x*x)/6 + (3*x*x*x*x*x)/40
}
