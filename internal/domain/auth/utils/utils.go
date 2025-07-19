package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"golang.org/x/crypto/bcrypt"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Claims represents the JWT claims for authentication
type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for the given claims
func GenerateJWT(c *gin.Context, username, userID, email string, expiresIn int64) (string, error) {
	// Get Secret Key from environment variable or configuration
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return "", NewJWTSecretNotSetError(c, "JWT secret key is not set")
	}

	// Validate input parameters later
	if username == "" || userID == "" || email == "" {
		return "", NewJWTClaimsInvalidError(c, "Invalid JWT claims provided")
	}

	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Minute)

	// Generate unique token ID
	tokenID, err := generateUUID(c)
	if err != nil {
		return "", NewUUIDGenerationFailedError(c, err)
	}

	// Create the JWT claims
	claims := &Claims{
		Username: username,
		UserID:   userID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "refyne-api",
			Subject:   userID,
			ID:        tokenID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, jwtErr := token.SignedString([]byte(secretKey))
	if jwtErr != nil {
		return "", NewJWTGenerationFailedError(c, jwtErr).WithOperation("GenerateJWT")
	}

	return tokenString, nil

}

// ValidateJWT validates the JWT token and returns the claims if valid
func ValidateJWT(c *gin.Context, tokenString string) (*Claims, *errors.AppError) {
	if tokenString == "" {
		return nil, NewJWTClaimsInvalidError(c, "JWT token is empty")
	}

	// Get Secret Key from environment variable or configuration
	secretKey := os.Getenv("JWT_SECRET_KEY")
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

// Password hashing and validation functions
func GenerateHash(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("invalid cost: %d", cost)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}
	return string(hash), nil
}

func CheckHash(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password does not match: %w", err)
	}
	return nil
}
