package auth

import "github.com/gin-gonic/gin"

// Request struct for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// Handler for this fuckass request
func (h *authHandler) Register(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User registered successfully",
	})
}

// Request struct for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Handler for user login
func (h *authHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User logged in successfully",
	})
}
