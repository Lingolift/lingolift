package middleware

import (
	_ "fmt"
	"time"

	"lingolift/api"
	"lingolift/config"
	"lingolift/errno"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AccessLogger
func AccessLogger(next echo.HandlerFunc) echo.HandlerFunc {
	rt := time.Now()
	return func(c echo.Context) error {

		// 1.是否需要忽略
		if IsIgnoreAuthRequest(c) {
			return next(c)
		}

		// 2.验证请求 header 参数是否合法
		if err := validatorHeaders(c); err.Code != errno.Success.Code {
			return api.ReturnError(c, err)
		}

		parameters := parseParamsFromRequest(c)

		err := next(c)

		var errCode, errMessage, rawErrMessage string

		var errResp api.ErrorResponse
		if err != nil {
			he, isOK := err.(*echo.HTTPError)
			if isOK {
				errResp = he.Message.(api.ErrorResponse)

				errCode = errResp.Error.Code
				errMessage = errResp.Error.Message

				if errResp.Error.RawErr != nil {
					rawErrMessage = errResp.Error.RawErr.Error()
				}
			}
		}

		config.AccessLogger.Info(errResp.Error.Message,
			zap.String("request_id", c.Request().Header.Get(config.HEADER_X_KSC_REQUEST_ID)),
			zap.String("method", c.Request().Method),
			zap.String("action", c.Request().URL.Path+c.Get(RequestActionCTX).(string)),
			zap.String("api_version", c.Get(RequestApiVersionCTX).(string)),
			zap.String("real_ip", c.Request().Header.Get(config.HEADER_X_KSC_REAL_IP)),
			zap.String("accept", c.Request().Header.Get(echo.HeaderAccept)),
			zap.String("content_type", c.Request().Header.Get(echo.HeaderContentType)),
			zap.String("user_agent", c.Request().Header.Get("User-Agent")),
			zap.String("account_id", c.Request().Header.Get(config.HEADER_X_KSC_ACCOUNT_ID)),
			zap.String("request_parameters", parameters),
			zap.Int("response_status_code", c.Response().Status),
			zap.String("error_code", errCode),
			zap.String("error_message", errMessage),
			zap.String("raw_error_message", rawErrMessage),
			zap.Duration("cost", time.Since(rt)),
		)

		if err != nil {
			return api.ReturnError(c, errResp.Error)
		}
		return err
	}
}
