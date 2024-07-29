package service

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// func (s *AuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
// 	user, err := s.userRepo.GetByEmail(ctx, email) //Search the database to check whether this email exists
// 	if err != nil {
// 		return "", err
// 	}

// 	if !user.CheckPassword(password) {
// 		return "", errors.New("Invalid Credentials")
// 	}

// 	//Generate JWT token
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": user.ID,
// 		"exp":     time.Now().Add(time.Hour * 24).Unix(),
// 	})

// 	tokenString, err := token.SignedString(s.jwtSecret) // takes the JWT token that we've created and populated with claims (like user ID and expiration time) and signs it using the secret key
// 	if err != nil {
// 		return "", nil
// 	}

// 	//Update the last login time
// 	s.userRepo.UpdateLastLogin(ctx, user.ID)

// 	return tokenString, nil
// }

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("invalid token")
}
