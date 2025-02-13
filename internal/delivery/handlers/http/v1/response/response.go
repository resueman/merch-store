package response

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	dto "github.com/resueman/merch-store/internal/api/v1"
)

func SendHandlerError(c echo.Context, httpCode int, message string) error {
	if e := c.JSON(httpCode, dto.ErrorResponse{Errors: &message}); e != nil {
		return fmt.Errorf("failed to send error response: %w", e)
	}

	return nil
}

func SendUsecaseError(c echo.Context, err error) error {
	httpCode, errMsg := getReturnHTTPCodeAndMessage(err)
	if e := c.JSON(httpCode, dto.ErrorResponse{Errors: &errMsg}); e != nil {
		return fmt.Errorf("failed to send error response: %w", e)
	}

	return nil
}

func SendNoContent(c echo.Context) error {
	if err := c.NoContent(http.StatusOK); err != nil {
		return fmt.Errorf("failed to send no content response: %w", err)
	}

	return nil
}

func SendOk(c echo.Context, data interface{}) error {
	if e := c.JSON(http.StatusOK, data); e != nil {
		return fmt.Errorf("failed to send success response: %w", e)
	}

	return nil
}
