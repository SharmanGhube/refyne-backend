package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key-for-unit-tests-minimum-32-bytes"

func setupGinTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

// --- GenerateJWT Tests ---

func TestGenerateJWT_ValidInputs(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	token, appErr := GenerateJWT(c, "testuser", "user-123", "test@example.com", 1, 15)

	require.Nil(t, appErr, "should not return error for valid inputs")
	assert.NotEmpty(t, token, "should return a non-empty token")

	// Verify the token is parseable
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsed.Valid)

	claims := parsed.Claims.(*Claims)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, 1, claims.TokenVersion)
	assert.Equal(t, "refyne-api", claims.Issuer)
	assert.Equal(t, "user-123", claims.Subject)
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), claims.ExpiresAt.Time, 5*time.Second)
}

func TestGenerateJWT_MissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	c := setupGinTestContext()

	token, appErr := GenerateJWT(c, "testuser", "user-123", "test@example.com", 1, 15)

	assert.Empty(t, token)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTSecretNotSet, appErr.Code)
}

func TestGenerateJWT_EmptyInputs(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	tests := []struct {
		name     string
		username string
		userID   string
		email    string
		expires  int64
	}{
		{"empty username", "", "user-123", "test@example.com", 15},
		{"empty userID", "testuser", "", "test@example.com", 15},
		{"empty email", "testuser", "user-123", "", 15},
		{"zero expiry", "testuser", "user-123", "test@example.com", 0},
		{"negative expiry", "testuser", "user-123", "test@example.com", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupGinTestContext()
			token, appErr := GenerateJWT(c, tt.username, tt.userID, tt.email, 1, tt.expires)

			assert.Empty(t, token)
			require.NotNil(t, appErr)
			assert.Equal(t, CodeJWTClaimsInvalid, appErr.Code)
		})
	}
}

// --- ValidateJWT Tests ---

func TestValidateJWT_ValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	// Generate → Validate round-trip
	token, appErr := GenerateJWT(c, "testuser", "user-123", "test@example.com", 1, 15)
	require.Nil(t, appErr)

	claims, appErr := ValidateJWT(c, token)

	require.Nil(t, appErr)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, 1, claims.TokenVersion)
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	// Create a token that expired 1 minute ago
	claims := &Claims{
		Username:     "testuser",
		UserID:       "user-123",
		Email:        "test@example.com",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-20 * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-20 * time.Minute)),
			Issuer:    "refyne-api",
			Subject:   "user-123",
			ID:        "expired-token-id",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	result, appErr := ValidateJWT(c, signedToken)

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTTokenExpired, appErr.Code)
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	// Sign with a different secret
	claims := &Claims{
		Username:     "testuser",
		UserID:       "user-123",
		Email:        "test@example.com",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "refyne-api",
			Subject:   "user-123",
			ID:        "wrong-secret-token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("completely-different-secret-key-here"))
	require.NoError(t, err)

	result, appErr := ValidateJWT(c, signedToken)

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTSigningMethodInvalid, appErr.Code)
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	result, appErr := ValidateJWT(c, "this.is.not.a.valid.jwt.token")

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTTokenInvalid, appErr.Code)
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	result, appErr := ValidateJWT(c, "")

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTClaimsInvalid, appErr.Code)
}

func TestValidateJWT_WrongIssuer(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	// Create a token with wrong issuer
	claims := &Claims{
		Username:     "testuser",
		UserID:       "user-123",
		Email:        "test@example.com",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wrong-issuer",
			Subject:   "user-123",
			ID:        "wrong-issuer-token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)

	result, appErr := ValidateJWT(c, signedToken)

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, CodeJWTClaimsInvalid, appErr.Code)
}

// --- GenerateTokenPair Tests ---

func TestGenerateTokenPair_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	pair, appErr := GenerateTokenPair(c, "testuser", "user-123", "test@example.com", 1)

	require.Nil(t, appErr)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, int64(15*60), pair.ExpiresIn) // 15 minutes in seconds
	assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)

	// Validate both tokens are valid
	accessClaims, appErr := ValidateJWT(c, pair.AccessToken)
	require.Nil(t, appErr)
	assert.Equal(t, "testuser", accessClaims.Username)

	refreshClaims, appErr := ValidateJWT(c, pair.RefreshToken)
	require.Nil(t, appErr)
	assert.Equal(t, "testuser", refreshClaims.Username)

	// Refresh token should have longer expiry than access token
	assert.True(t, refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time))
}

// --- ValidateAndExtractToken Tests ---

func TestValidateAndExtractToken_Valid(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	c := setupGinTestContext()

	token, appErr := GenerateJWT(c, "testuser", "user-123", "test@example.com", 1, 15)
	require.Nil(t, appErr)

	claims, err := ValidateAndExtractToken(token)

	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestValidateAndExtractToken_EmptyToken(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	claims, err := ValidateAndExtractToken("")

	assert.Nil(t, claims)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateAndExtractToken_MissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	claims, err := ValidateAndExtractToken("some.token.here")

	assert.Nil(t, claims)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret")
}
