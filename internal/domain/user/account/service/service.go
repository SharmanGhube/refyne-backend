package service

import (
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type UserAccountService interface {
}

type userAccountService struct {
	logger      *zap.Logger
	serviceName string
}

func NewUserAccountService() UserAccountService {
	return &userAccountService{
		logger:      logging.GetServiceLogger("UserAccountService"),
		serviceName: "UserAccountService",
	}
}
