package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
)

type authUsecase struct {
	userRepo        repo.User
	secretKey       string
	tokenTTL        time.Duration
	passwordManager PasswordManager
}

func NewAuthUsecase(userRepo repo.User, passwordManager PasswordManager) *authUsecase {
	return &authUsecase{userRepo: userRepo, passwordManager: passwordManager}
}

type PasswordManager interface {
	HashPassword(password string) string
	ComparePassword(password, hash string) bool
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID int
}

var emptyClaims = model.Claims{}

func (u *authUsecase) GenerateToken(ctx context.Context, input model.AuthRequestInput) (string, error) {
	user, err := u.userRepo.GetUserByUsername(ctx, input.Username)
	if err == nil {
		if !u.passwordManager.ComparePassword(input.Password, user.Hash) {
			return "", apperrors.ErrInvalidPassword
		}

		return u.generateToken(model.Claims{UserID: user.ID})
	}

	if !errors.Is(err, apperrors.ErrUserNotFound) {
		return "", err
	}

	userID, err := u.registerUser(ctx, input) // Write!!!!!!!!!!!!
	if err != nil {
		return "", err
	}

	return u.generateToken(model.Claims{UserID: userID})
}

func (u *authUsecase) generateToken(claims model.Claims) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(u.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: claims.UserID,
	})

	tokenString, err := token.SignedString([]byte(u.secretKey))
	if err != nil {
		return "", apperrors.ErrGenerateToken
	}

	return tokenString, nil
}

func (u *authUsecase) registerUser(ctx context.Context, input model.AuthRequestInput) (int, error) {
	newUser := &entity.CreateUserInput{
		Username: input.Username,
		Hash:     u.passwordManager.HashPassword(input.Password),
	}

	userID, err := u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (u *authUsecase) ParseToken(ctx context.Context, tokenString string) (model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %s", token.Header["alg"])
		}

		return []byte(u.secretKey), nil
	})

	if validationErr, ok := err.(*jwt.ValidationError); ok {
		if validationErr.Errors == jwt.ValidationErrorExpired {
			return emptyClaims, apperrors.ErrTokenExpired
		}
	}

	if err != nil {
		return emptyClaims, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return emptyClaims, apperrors.ErrInvalidToken
	}

	return model.Claims{UserID: claims.UserID}, nil
}
