package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Claims represents the JWT claims for authentication
type Claims struct {
	Username     string `json:"username"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(c *gin.Context, username, userID, email string, tokenVersion int) (*TokenPair, *errors.AppError) {
	// Access Token valid for 15 minutes
	accessToken, appErr := GenerateJWT(c, username, userID, email, tokenVersion, 15)
	if appErr != nil {
		return nil, appErr
	}

	// Refresh Token valid for 7 days (10080 minutes)
	refreshToken, appErr := GenerateJWT(c, username, userID, email, tokenVersion, 10080)
	if appErr != nil {
		return nil, appErr
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // in seconds
	}, nil
}

// GenerateJWT generates a JWT token for the given claims
func GenerateJWT(c *gin.Context, username, userID, email string, tokenVersion int, expiresIn int64) (string, *errors.AppError) {
	// Get Secret from env
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", NewJWTSecretNotSetError(c, "JWT secret is not set in environment variables")
	}

	// Validate input parameters
	if username == "" || userID == "" || email == "" || expiresIn <= 0 {
		return "", NewJWTClaimsInvalidError(c, "Invalid claims provided for JWT generation")
	}

	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Minute)

	// Generate unique token ID
	tokenID, appErr := generateUUID(c)
	if appErr != nil {
		return "", appErr
	}

	// Create the JWT claims
	claims := &Claims{
		Username:     username,
		UserID:       userID,
		Email:        email,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        tokenID,
			Issuer:    "refyne-api",
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", NewJWTGenerationError(c, "Failed to sign JWT token: "+err.Error())
	}

	return signedToken, nil
}

// ValidateJWT validates the JWT token and returns the claims if valid
func ValidateJWT(c *gin.Context, tokenString string) (*Claims, *errors.AppError) {
	if tokenString == "" {
		return nil, NewJWTClaimsInvalidError(c, "JWT token is empty")
	}

	// Get Secret Key from environment variable or configuration
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, NewJWTSecretNotSetError(c, "JWT secret key is not set")
	}

	// Parse and validate the token
	claims := &Claims{}
	token, jwtErr := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method: %v", method)
		}
		return []byte(secretKey), nil
	})
	if jwtErr != nil {
		// Handler Specific JWT errors
		if ve, ok := jwtErr.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, NewJWTTokenInvalidError(c, "Malformed JWT token")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, NewJWTTokenExpiredError(c, "JWT token has expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, NewJWTTokenInvalidError(c, "JWT token is not valid yet")
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, NewJWTSigningMethodInvalidError(c, "JWT token signature is invalid")
			}
		}

		return nil, NewJWTTokenInvalidError(c, "Failed to parse JWT token").WithContext(
			"error", jwtErr.Error(),
		)
	}

	if !token.Valid {
		return nil, NewJWTTokenInvalidError(c, "JWT token is invalid")
	}

	// Validate claims
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, NewJWTTokenExpiredError(c, "JWT token has expired")
	}

	// Validate other claims if needed
	if claims.Issuer != "refyne-api" {
		return nil, NewJWTClaimsInvalidError(c, "Invalid JWT issuer")
	}
	if claims.Subject == "" {
		return nil, NewJWTClaimsInvalidError(c, "Invalid JWT subject")
	}

	// Return the validated claims
	return claims, nil

}

// Helpers
func generateUUID(c *gin.Context) (string, *errors.AppError) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", NewUUIDGenerationFailedError(c, err)
	}
	return newUUID.String(), nil
}

// ValidateAndExtractToken validates JWT token and extracts claims
func ValidateAndExtractToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// Get secret from environment
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
