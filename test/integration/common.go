package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	v1 "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/account"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/auth"
	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/operation"
	"github.com/resueman/merch-store/internal/delivery/middleware"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/pkg/db/postgres"
	"github.com/resueman/merch-store/pkg/password"
	"github.com/stretchr/testify/assert"
)

var (
	router           *echo.Echo
	authHandler      *auth.AuthHandler
	operationHandler *operation.OperationHandler
	accountHandler   *account.AccountHandler
	dbClient         db.Client
	authMiddleware   *middleware.AuthMiddleware
)

func setup() {
	var err error
	dbClient, err = postgres.NewPostgresClient(context.Background(),
		"postgres://user:password@localhost:5433/store?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	txManager := postgres.NewTxManager(dbClient, time.Second*10, 3)
	repositories := repo.NewRepositories(dbClient)
	passwordManager := password.NewPasswordManager("1234567890")
	tokenTTL := time.Minute * 15
	usecases := usecase.NewUsecase(repositories, txManager, passwordManager, "secret", tokenTTL)

	router = echo.New()
	authMiddleware = middleware.NewAuthMiddleware(usecases)
	authHandler = auth.NewAuthHandler(router, usecases)
	operationHandler = operation.NewOperationHandler(router, usecases)
	accountHandler = account.NewAccountHandler(router, usecases)
}

func cleanup() {
	dbClient.Primary().Exec(context.Background(), db.Query{QueryRaw: "DELETE FROM purchase_operations"})
	dbClient.Primary().Exec(context.Background(), db.Query{QueryRaw: "DELETE FROM transfer_operations"})
	dbClient.Primary().Exec(context.Background(), db.Query{QueryRaw: "DELETE FROM operations"})
	dbClient.Primary().Exec(context.Background(), db.Query{QueryRaw: "DELETE FROM accounts"})
	dbClient.Primary().Exec(context.Background(), db.Query{QueryRaw: "DELETE FROM users"})

	dbClient.Close()
	router.Close()
}

func authUser(t *testing.T, username, password string, expectedStatus int) string {
	t.Helper()

	reqInput := v1.AuthRequest{Username: username, Password: password}
	body, err := json.Marshal(reqInput)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	ctx := router.NewContext(request, recorder)

	err = authHandler.Auth(ctx)
	assert.NoError(t, err)

	assert.Equal(t, expectedStatus, recorder.Code)

	var response v1.AuthResponse

	err = json.Unmarshal([]byte(recorder.Body.String()), &response)
	if err != nil {
		t.Fatal(err)
	}

	return *response.Token
}

func buyItem(t *testing.T, token string, item string, expectedStatus int) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/api/buy/"+item, nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	ctx := router.NewContext(request, recorder)
	ctx.SetParamNames("item")
	ctx.SetParamValues(item)

	err := authMiddleware.AuthMiddleware(operationHandler.BuyItem)(ctx)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedStatus, recorder.Code)
	}
}

func sendCoin(t *testing.T, token string, toUser string, amount int, expectedStatus int) {
	t.Helper()

	requestInput := v1.SendCoinRequest{ToUser: toUser, Amount: amount}
	body, err := json.Marshal(requestInput)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sendCoin",
		bytes.NewReader(body))

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	ctx := router.NewContext(request, recorder)

	err = authMiddleware.AuthMiddleware(operationHandler.SendCoin)(ctx)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedStatus, recorder.Code)
	}
}

func getUserInfo(t *testing.T, token string, expectedStatus int, expected *v1.InfoResponse) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/api/account", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	ctx := router.NewContext(request, recorder)

	err := authMiddleware.AuthMiddleware(accountHandler.GetInfo)(ctx)
	assert.NoError(t, err)

	actual := &v1.InfoResponse{}
	if err := json.Unmarshal(recorder.Body.Bytes(), actual); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedStatus, recorder.Code)
	assert.True(t, reflect.DeepEqual(expected, actual))
}
