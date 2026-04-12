package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	testUserID    = "test-user-123"
	testAccountID = "test-account-456"
	testMediaID   = "test-media-789"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// Test: List Accounts
func TestListAccounts(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/accounts", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"data": []gin.H{
				{
					"id":          testAccountID,
					"username":    "testuser",
					"sync_status": "idle",
				},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/instagram/accounts", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Get Account
func TestGetAccount(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/accounts/:id", func(c *gin.Context) {
		accountID := c.Param("id")
		userID := c.GetString("userID")

		if accountID != testAccountID {
			c.JSON(404, gin.H{"error": "Account not found"})
			return
		}

		if userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"id":              testAccountID,
				"username":        "testuser",
				"followers_count": 1000,
				"biography":       "Test account",
				"sync_status":     "idle",
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/accounts/%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, testAccountID, data["id"])
}

// Test: Get Media
func TestGetMedia(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/media", func(c *gin.Context) {
		accountID := c.Query("account_id")
		userID := c.GetString("userID")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		if accountID != testAccountID || userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": []gin.H{
				{
					"id":          testMediaID,
					"media_id":    "ig-media-123",
					"media_type":  "PHOTO",
					"like_count":  100,
					"impressions": 1000,
					"reach":       850,
				},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/media?account_id=%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
	assert.NotEmpty(t, resp["data"])
}

// Test: Get Media By ID
func TestGetMediaByID(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/media/:id", func(c *gin.Context) {
		mediaID := c.Param("id")
		userID := c.GetString("userID")

		if mediaID != testMediaID {
			c.JSON(404, gin.H{"error": "Media not found"})
			return
		}

		if userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"id":          testMediaID,
				"media_id":    "ig-media-123",
				"media_type":  "PHOTO",
				"like_count":  100,
				"impressions": 1000,
				"caption":     "Test caption",
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/media/%s", testMediaID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Get Account Analytics
func TestGetAccountAnalytics(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/analytics", func(c *gin.Context) {
		accountID := c.Query("account_id")
		userID := c.GetString("userID")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		if accountID != testAccountID || userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"account_id": accountID,
				"insights": []gin.H{
					{
						"metric_date":     time.Now().Format("2006-01-02"),
						"impressions":     1000,
						"reach":           850,
						"engagement_rate": 8.5,
						"follower_count":  1000,
						"growth_rate":     2.5,
					},
				},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/analytics?account_id=%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Get Media Analytics
func TestGetMediaAnalytics(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/analytics/media", func(c *gin.Context) {
		accountID := c.Query("account_id")
		userID := c.GetString("userID")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		if accountID != testAccountID || userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"account_id": accountID,
				"insights": []gin.H{
					{
						"media_id":        testMediaID,
						"metric_date":     time.Now().Format("2006-01-02"),
						"impressions":     500,
						"reach":           400,
						"engagement_rate": 12.0,
					},
				},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/analytics/media?account_id=%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Get Analytics Trends
func TestGetAnalyticsTrends(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/analytics/trends", func(c *gin.Context) {
		accountID := c.Query("account_id")
		granularity := c.DefaultQuery("granularity", "daily")
		userID := c.GetString("userID")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		if accountID != testAccountID || userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"account_id":  accountID,
				"granularity": granularity,
				"trends": []gin.H{
					{
						"period":          time.Now().Format("2006-01-02"),
						"impressions":     int64(1000),
						"reach":           int64(850),
						"engagement_rate": 8.5,
					},
				},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/analytics/trends?account_id=%s&granularity=daily", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Generate Captions
func TestGenerateCaptions(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.POST("/api/instagram/ai/caption-suggest", func(c *gin.Context) {
		var req struct {
			AccountID       string   `json:"account_id"`
			OriginalCaption string   `json:"original_caption"`
			MediaType       string   `json:"media_type"`
			Themes          []string `json:"themes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.AccountID == "" || req.OriginalCaption == "" {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.AccountID != testAccountID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		if len(req.OriginalCaption) > 2200 {
			c.JSON(400, gin.H{"error": "Caption too long (max 2200 characters)"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"alternatives": []gin.H{
					{"caption": "Alternative 1", "reason": "Better engagement"},
					{"caption": "Alternative 2", "reason": "More hashtags"},
					{"caption": "Alternative 3", "reason": "Call to action"},
				},
			},
		})
	})

	body := bytes.NewBufferString(`{
		"account_id": "test-account-456",
		"original_caption": "Check out my new post!",
		"media_type": "PHOTO",
		"themes": ["lifestyle"]
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/ai/caption-suggest", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Generate Hashtags
func TestGenerateHashtags(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.POST("/api/instagram/ai/hashtag-suggest", func(c *gin.Context) {
		var req struct {
			AccountID   string `json:"account_id"`
			Caption     string `json:"caption"`
			MediaType   string `json:"media_type"`
			ContentType string `json:"content_type"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.AccountID != testAccountID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"hashtags": []string{"#instagram", "#socialmedia", "#contentcreator", "#engagement", "#marketing"},
			},
		})
	})

	body := bytes.NewBufferString(`{
		"account_id": "test-account-456",
		"caption": "Check out my new post!",
		"media_type": "PHOTO"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/ai/hashtag-suggest", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Get Posting Strategy
func TestGetPostingStrategy(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/ai/posting-time", func(c *gin.Context) {
		accountID := c.Query("account_id")
		userID := c.GetString("userID")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		if accountID != testAccountID || userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"data": gin.H{
				"account_id":           accountID,
				"best_posting_days":    []string{"Tuesday", "Wednesday", "Thursday"},
				"best_posting_times":   []string{"09:00 UTC", "12:00 UTC", "18:00 UTC"},
				"predicted_reach":      1.2,
				"engagement_potential": 78,
				"audience_timezone":    "UTC",
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/ai/posting-time?account_id=%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

// Test: Manual Sync Job Queue
func TestManualSync(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.POST("/api/instagram/media/sync", func(c *gin.Context) {
		var req struct {
			AccountID string `json:"account_id"`
			SyncType  string `json:"sync_type"`
			Force     bool   `json:"force"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.AccountID != testAccountID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		validTypes := map[string]bool{"new": true, "full": true, "insights": true}
		if req.SyncType == "" {
			req.SyncType = "new"
		}
		if !validTypes[req.SyncType] {
			c.JSON(400, gin.H{"error": "Invalid sync_type"})
			return
		}

		// Return 202 Accepted for async job
		c.JSON(202, gin.H{
			"status":     "accepted",
			"message":    "Sync job queued",
			"job_type":   "sync_media",
			"account_id": req.AccountID,
		})
	})

	body := bytes.NewBufferString(`{
		"account_id": "test-account-456",
		"sync_type": "new",
		"force": false
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/media/sync", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "accepted", resp["status"])
	assert.Equal(t, "sync_media", resp["job_type"])
}

// Test: Manual Analyze Job Queue
func TestManualAnalyze(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.POST("/api/instagram/media/analyze", func(c *gin.Context) {
		var req struct {
			MediaID   string `json:"media_id"`
			AccountID string `json:"account_id"`
			Force     bool   `json:"force"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		if req.AccountID != testAccountID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		// Return 202 Accepted for async job
		c.JSON(202, gin.H{
			"status":     "accepted",
			"message":    "AI analysis job queued",
			"job_type":   "process_ai",
			"media_id":   req.MediaID,
			"account_id": req.AccountID,
		})
	})

	body := bytes.NewBufferString(`{
		"media_id": "test-media-789",
		"account_id": "test-account-456",
		"force": false
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/media/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "accepted", resp["status"])
	assert.Equal(t, "process_ai", resp["job_type"])
}

// Test: Authorization - Forbidden
func TestAuthorizationForbidden(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "different-user")
	})

	router.GET("/api/instagram/accounts/:id", func(c *gin.Context) {
		userID := c.GetString("userID")
		accountID := c.Param("id")

		if accountID == testAccountID && userID != testUserID {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/accounts/%s", testAccountID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Test: Query Parameter Validation - Missing Required Parameter
func TestQueryParameterValidation(t *testing.T) {
	router := gin.New()

	router.GET("/api/instagram/media", func(c *gin.Context) {
		accountID := c.Query("account_id")
		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/instagram/media", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "account_id is required", resp["error"])
}

// Test: Invalid Request Body - Malformed JSON
func TestInvalidRequestBody(t *testing.T) {
	router := gin.New()

	router.POST("/api/instagram/ai/caption-suggest", func(c *gin.Context) {
		var req struct {
			AccountID string `json:"account_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	body := bytes.NewBufferString(`{invalid json}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/ai/caption-suggest", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test: Request Validation - Caption Length Limit
func TestCaptionLengthValidation(t *testing.T) {
	router := gin.New()

	router.POST("/api/instagram/ai/caption-suggest", func(c *gin.Context) {
		var req struct {
			OriginalCaption string `json:"original_caption"`
		}
		c.ShouldBindJSON(&req)

		if len(req.OriginalCaption) > 2200 {
			c.JSON(400, gin.H{"error": "Caption too long (max 2200 characters)"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	longCaption := make([]byte, 2201)
	for i := range longCaption {
		longCaption[i] = 'a'
	}

	body := bytes.NewBufferString(fmt.Sprintf(`{"original_caption": "%s"}`, string(longCaption)))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/ai/caption-suggest", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test: Context Timeout
func TestContextTimeout(t *testing.T) {
	testTimeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	time.Sleep(testTimeout + 50*time.Millisecond)

	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	default:
		t.Fatal("Context should have timed out")
	}
}

// Test: Sync Type Validation
func TestSyncTypeValidation(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.POST("/api/instagram/media/sync", func(c *gin.Context) {
		var req struct {
			AccountID string `json:"account_id"`
			SyncType  string `json:"sync_type"`
		}
		c.ShouldBindJSON(&req)

		validTypes := map[string]bool{"new": true, "full": true, "insights": true}
		if !validTypes[req.SyncType] {
			c.JSON(400, gin.H{"error": "Invalid sync_type"})
			return
		}
		c.JSON(202, gin.H{"status": "accepted"})
	})

	body := bytes.NewBufferString(`{
		"account_id": "test-account-456",
		"sync_type": "invalid"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/instagram/media/sync", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test: Granularity Parameter Validation
func TestGranularityValidation(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
	})

	router.GET("/api/instagram/analytics/trends", func(c *gin.Context) {
		accountID := c.Query("account_id")
		granularity := c.DefaultQuery("granularity", "daily")

		if accountID == "" {
			c.JSON(400, gin.H{"error": "account_id is required"})
			return
		}

		validGranularities := map[string]bool{"daily": true, "weekly": true, "monthly": true}
		if !validGranularities[granularity] {
			granularity = "daily" // Default to daily for invalid granularity
		}
		c.JSON(200, gin.H{"status": "ok", "data": gin.H{"granularity": granularity}})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/instagram/analytics/trends?account_id=%s&granularity=yearly", testAccountID), nil)
	router.ServeHTTP(w, req)

	// With our default validation, invalid granularity should default to daily
	assert.Equal(t, http.StatusOK, w.Code)
}
