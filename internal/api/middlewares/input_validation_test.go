package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Content-Type Validation ---

func TestInputValidation_ContentType_JSON_Accepted(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = 15
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInputValidation_ContentType_JSONWithCharset_Accepted(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.ContentLength = 15
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInputValidation_ContentType_TextHTML_Rejected(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`<html>bad</html>`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "text/html")
	req.ContentLength = 16
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

func TestInputValidation_ContentType_EmptyBody_Skipped(t *testing.T) {
	// POST with no body (ContentLength=0) should skip Content-Type check
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	// No Content-Type header, but body is empty
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInputValidation_ContentType_GET_Ignored(t *testing.T) {
	// GET requests should never check Content-Type
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Path Parameter Sanitization ---

func TestInputValidation_PathParam_TraversalRejected(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/files/:name", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"file": c.Param("name")})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/..%2F..%2Fetc%2Fpasswd", nil)
	r.ServeHTTP(w, req)

	// Gin URL-decodes params, so "../" becomes ".." which triggers check
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInputValidation_PathParam_NullByteRejected(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/files/:name", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"file": c.Param("name")})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/test%2500file", nil)
	r.ServeHTTP(w, req)

	// "%00" in URL-decoded param triggers null byte check
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInputValidation_PathParam_SafeValueAccepted(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/abc-123-def", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- containsDangerousPathChars unit tests ---

func TestContainsDangerousPathChars(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"path traversal double dot", "../etc/passwd", true},
		{"null byte encoded", "file%00.txt", true},
		{"literal null byte", "file\x00.txt", true},
		{"control char (bell)", "file\x07name", true},
		{"normal value", "my-file-123", false},
		{"single dot (ok)", "file.txt", false},
		{"tab (allowed)", "file\tname", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, containsDangerousPathChars(tt.value))
		})
	}
}

// --- ValidateRequestSize ---

func TestValidateRequestSize_UnderLimit(t *testing.T) {
	r := gin.New()
	r.Use(ValidateRequestSize(1024))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"small":"body"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.ContentLength = 16
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequestSize_OverLimit(t *testing.T) {
	r := gin.New()
	r.Use(ValidateRequestSize(10))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"large":"this body is definitely over ten bytes"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.ContentLength = 50
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

// --- Skip Validation ---

func TestInputValidation_SkipsHealthEndpoint(t *testing.T) {
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
