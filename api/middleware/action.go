package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"lingolift/config"
	"lingolift/errno"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

const (
	// 记录请求的 Action
	RequestActionCTX     = "Action"
	RequestApiVersionCTX = "Version"

	// 记录请求开始时间
	RequestStartTimeCTX = "RequestStartTime"
)

var (
	//请求必须携带的 header 信息
	requiredHeaders = []string{
		config.HEADER_X_KSC_REQUEST_ID,
		config.HEADER_X_KSC_ACCOUNT_ID,
	}
)

// ParseAction 解析请求中的 `Action` 字段
// func ParseAction(c echo.Context) (string, error) {
// 	actionInterface := c.Get(RequestActionCTX)
// 	action, ok := actionInterface.(string)
// 	if !ok || action == "" {
// 		return "", api.ErrActionRequest(c, action)
// 	}
// 	return action, nil
// }

// parseParamsFromRequest 解析请求中的参数
func parseParamsFromRequest(c echo.Context) string {
	switch c.Request().Method {
	case http.MethodPost:
		return ParseParamsFromBody(c)
	default:
		return ParseParamsFromQuery(c)
	}
}

// ParseParamsFromBody 从Body中解析相关参数
func ParseParamsFromBody(c echo.Context) string {
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	params := string(bodyBytes)

	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	params = strings.Replace(params, " ", "", -1)
	params = strings.Replace(params, "\n", "", -1)

	action := c.QueryParam("Action")
	apiVersion := c.QueryParam("Version")
	if action == "" {
		actionInterface := gjson.Get(params, "Action")
		action = actionInterface.String()

		versionInterface := gjson.Get(params, "Version")
		apiVersion = versionInterface.String()
	}

	c.Set(RequestActionCTX, action)
	c.Set(RequestApiVersionCTX, apiVersion)

	if len(params) > 4096 {
		params = params[:4096] + "..."
	}

	return params
}

// ParseParamsFromQuery 从URL解析相关参数
func ParseParamsFromQuery(c echo.Context) string {
	params := c.Request().URL.RawQuery
	if len(params) > 512 {
		params = params[:512] + "..."
	}

	action := c.QueryParam("Action")
	apiVersion := c.QueryParam("Version")

	c.Set(RequestActionCTX, action)
	c.Set(RequestApiVersionCTX, apiVersion)

	return params
}

// IsIgnoreAuthRequest
func IsIgnoreAuthRequest(c echo.Context) bool {
	if strings.Contains(c.Request().URL.Path, "debug/pprof") ||
		strings.Contains(c.Request().URL.Path, "health") ||
		strings.Contains(c.Request().URL.Path, "metrics") {
		return true
	}

	return false
}

// validatorHeaders
func validatorHeaders(c echo.Context) errno.Err {
	// for _, k := range requiredHeaders {
	// 	v := c.Request().Header.Get(k)
	// 	if v == "" {
	// 		return errno.ErrMissingHeader.WithFmt(k)
	// 	}
	// }
	return *errno.Success
}
