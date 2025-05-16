package api

import (
	"fmt"
	"net/http"

	"lingolift/config"
	"lingolift/errno"

	"github.com/labstack/echo/v4"
)

// ErrorResponse standard error response format.
type ErrorResponse struct {
	RequestID string    `json:"RequestID" xml:"RequestID"`
	Error     errno.Err `json:"Error" xml:"Error"`
}

// ErrURLRequest
func ErrURLRequest(c echo.Context) error {
	return echo.NewHTTPError(404, ErrorResponse{
		RequestID: c.Request().Header.Get(config.HEADER_X_KSC_REQUEST_ID),
		Error:     errno.ErrNoSuchEntity.WithRawErr(fmt.Errorf("Not found request URL.")),
	})
}

// ReturnError return errors in a unified format.
func ReturnError(c echo.Context, e errno.Err) error {
	resp := ErrorResponse{
		RequestID: c.Request().Header.Get(config.HEADER_X_KSC_REQUEST_ID),
		Error:     e,
	}

	return echo.NewHTTPError(e.HTTPCode, resp)
}

// SuccessResponse
type SuccessResponse struct {
	RequestID string `json:"RequestID"`
}

// ReturnSuccess
func ReturnSuccess(c echo.Context) error {
	resp := SuccessResponse{
		RequestID: c.Request().Header.Get(config.HEADER_X_KSC_REQUEST_ID),
	}

	return Return(c, resp)
}

// Return
func Return(c echo.Context, data interface{}) error {
	acceptType := c.Request().Header.Get("Accept")
	if acceptType == echo.MIMEApplicationXML {
		return c.XML(http.StatusOK, data)
	}
	return c.JSON(http.StatusOK, data)
}
