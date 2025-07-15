package auth

type AuthService interface {
}

type authService struct {
}

func NewAuthService() AuthService {
	return &authService{}
}
