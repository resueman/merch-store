package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/resueman/merch-store/internal/repo/repoerrors"
	"github.com/resueman/merch-store/internal/usecase/apperrors"
	"github.com/resueman/merch-store/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestAuthUsecase_GenerateTokenWithRegistration_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUser(ctrl)
	passwordManager := mocks.NewMockPasswordManager(ctrl)

	authRequestInput := model.AuthRequestInput{Username: "test", Password: "password"}
	hash, userID := "hash", 123
	mock := func() {
		userRepo.EXPECT().
			GetUserByUsername(gomock.Any(), authRequestInput.Username).
			Return(nil, repoerrors.ErrNotFound)

		passwordManager.EXPECT().
			HashPassword(authRequestInput.Password).
			Return(hash)

		userRepo.EXPECT().
			CreateUser(gomock.Any(), &entity.CreateUserInput{
				Username: authRequestInput.Username,
				Hash:     hash,
			}).
			Return(userID, nil)
	}

	mock()

	secretKey, tokenTTL := "secret", time.Minute*15
	authUsecase := NewAuthUsecase(userRepo, passwordManager, secretKey, tokenTTL)
	token, err := authUsecase.GenerateToken(context.Background(), authRequestInput)

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestAuthUsecase_GenerateTokenWithRegistration_ErrorRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUser(ctrl)
	passwordManager := mocks.NewMockPasswordManager(ctrl)

	registerUserErr := errors.New("error")
	authRequestInput := model.AuthRequestInput{Username: "test", Password: "password"}
	hash := "hash"
	mock := func() {
		userRepo.EXPECT().
			GetUserByUsername(gomock.Any(), authRequestInput.Username).
			Return(nil, repoerrors.ErrNotFound)

		passwordManager.EXPECT().
			HashPassword(authRequestInput.Password).
			Return(hash)

		userRepo.EXPECT().
			CreateUser(gomock.Any(), &entity.CreateUserInput{
				Username: authRequestInput.Username,
				Hash:     hash,
			}).
			Return(0, registerUserErr)
	}

	mock()

	authUsecase := NewAuthUsecase(userRepo, passwordManager, "secret", time.Minute*15)
	_, err := authUsecase.GenerateToken(context.Background(), authRequestInput)

	require.ErrorIs(t, err, registerUserErr)
}

func TestAuthUsecase_GenerateTokenForExistingUser_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUser(ctrl)
	passwordManager := mocks.NewMockPasswordManager(ctrl)

	secretKey := "secret"
	tokenTTL := time.Minute * 15

	authUsecase := NewAuthUsecase(userRepo, passwordManager, secretKey, tokenTTL)

	userRepo.EXPECT().
		GetUserByUsername(gomock.Any(), gomock.Any()).
		Return(&entity.User{ID: 1, Username: "test", Hash: "hash"}, nil)

	passwordManager.EXPECT().
		ComparePassword(gomock.Any(), gomock.Any()).
		Return(true)

	token, err := authUsecase.GenerateToken(context.Background(), model.AuthRequestInput{
		Username: "test",
		Password: "password",
	})

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestAuthUsecase_GenerateTokenForExistingUser_IncorrectPasswordError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUser(ctrl)
	passwordManager := mocks.NewMockPasswordManager(ctrl)

	authRequestInput := model.AuthRequestInput{
		Username: "test",
		Password: "password",
	}

	mock := func() {
		userRepo.EXPECT().
			GetUserByUsername(gomock.Any(), authRequestInput.Username).
			Return(&entity.User{ID: 1, Username: "test", Hash: "hash"}, nil)

		passwordManager.EXPECT().
			ComparePassword(gomock.Any(), gomock.Any()).
			Return(false)
	}

	mock()

	secretKey, tokenTTL := "secret", time.Minute*15
	authUsecase := NewAuthUsecase(userRepo, passwordManager, secretKey, tokenTTL)
	_, err := authUsecase.GenerateToken(context.Background(), authRequestInput)

	require.ErrorIs(t, err, apperrors.ErrInvalidPassword)
}

func TestAuthUsecase_GenerateToken_ErrorGettingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUser(ctrl)
	passwordManager := mocks.NewMockPasswordManager(ctrl)

	authRequestInput := model.AuthRequestInput{
		Username: "test",
		Password: "password",
	}

	errorGettingUser := errors.New("error")
	mock := func() {
		userRepo.EXPECT().
			GetUserByUsername(gomock.Any(), authRequestInput.Username).
			Return(nil, errorGettingUser)
	}

	mock()

	authUsecase := NewAuthUsecase(userRepo, passwordManager, "secret", time.Minute*15)
	_, err := authUsecase.GenerateToken(context.Background(), authRequestInput)

	require.ErrorIs(t, err, errorGettingUser)
}
