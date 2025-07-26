package account

import (
	"github.com/gin-gonic/gin"
	accountService "github.com/refynehq/refyne-backend/internal/domain/user/account/service"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type UserAccountHandler interface {
	UpdatePassword(c *gin.Context)
}

type UserAccountHandlerImpl struct {
	name   string
	logger *zap.Logger

	// Service Dependencies
	accountService accountService.UserAccountService
}

func NewUserAccountHandler(accountService accountService.UserAccountService) UserAccountHandler {
	return &UserAccountHandlerImpl{
		name:           "UserAccountHandler",
		logger:         logging.GetHandlerLogger("userAccountHandler"),
		accountService: accountService,
	}
}

func (h *UserAccountHandlerImpl) UpdatePassword(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Update Password endpoint hit",
	})
}
