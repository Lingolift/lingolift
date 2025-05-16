package middleware

import (
	"fmt"
	"runtime"

	"lingolift/api"
	cfg "lingolift/config"
	"lingolift/errno"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type (
	// RecoverConfig defines the config for Recover middleware.
	RecoverConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Size of the stack to be printed.
		// Optional. Default value 4KB.
		StackSize int `json:"stack_size"`

		// DisableStackAll disables formatting stack traces of all other goroutines
		// into buffer after the trace for the current goroutine.
		// Optional. Default value false.
		DisableStackAll bool `json:"disable_stack_all"`

		// DisablePrintStack disables printing stack trace.
		// Optional. Default value as false.
		DisablePrintStack bool `json:"disable_print_stack"`
	}
)

var (
	// DefaultRecoverConfig is the default Recover middleware config.
	DefaultRecoverConfig = RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover() echo.MiddlewareFunc {
	return RecoverWithConfig(DefaultRecoverConfig)
}

// RecoverWithConfig returns a Recover middleware with config.
// See: `Recover()`.
func RecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					var err error
					switch r := r.(type) {
					case error:
						err = r
					default:
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						cfg.AccessLogger.Error("[PANIC RECOVER]",
							zap.String("request_id", c.Request().Header.Get(cfg.HEADER_X_KSC_REQUEST_ID)),
							zap.String("method", c.Request().Method),
							zap.String("action", c.Request().URL.Path+c.Get(RequestActionCTX).(string)),
							zap.String("api_version", c.Get(RequestApiVersionCTX).(string)),
							zap.String("real_ip", c.Request().Header.Get(cfg.HEADER_X_KSC_REAL_IP)),
							zap.String("accept", c.Request().Header.Get(echo.HeaderAccept)),
							zap.String("content_type", c.Request().Header.Get(echo.HeaderContentType)),
							zap.String("user_agent", c.Request().Header.Get("User-Agent")),
							zap.String("account_id", c.Request().Header.Get(cfg.HEADER_X_KSC_ACCOUNT_ID)),
							zap.String("raw_error_message", string(stack[:length])),
						)
					}

					api.ReturnError(c, errno.ErrPanicException.WithRawErr(err))
				}
			}()

			return next(c)
		}
	}
}
