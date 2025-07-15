package auth

import "github.com/gin-gonic/gin"

func (h *authHandler) Register(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User registered successfully",
	})
}

func (h *authHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User logged in successfully",
	})
}
