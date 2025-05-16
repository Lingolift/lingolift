package errno

import "net/http"

// Success indicates a successful operation with no errors.
var Success = &Err{HTTPCode: http.StatusOK, Code: "Success", Message: "Success"}

// Server Errors - these are internal system errors that occur during processing.
var (
	// ErrInternalRequest indicates that the internal service request has failed.
	ErrInternalRequest = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "InternalError",
		Message:  "Internal service request failed.",
	}

	// ErrInternalServer indicates that there was an internal server error.
	ErrInternalServer = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "ServiceUnavailable",
		Message:  "Internal server error.",
	}

	// ErrServiceTimeout indicates that the internal service is unavailable because of a timeout.
	ErrServiceTimeout = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "ServiceTimeout",
		Message:  "Internal service is unavailable because of time out.",
	}

	// ErrExecTimeout indicates that the execution has exceeded the allowed timeout period.
	ErrExecTimeout = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "ExecTimeout",
		Message:  "Service execution timed out.",
	}

	// ErrDatabase indicates an exception occurred during database operations.
	ErrDatabase = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "ServiceException",
		Message:  "Data operation exception.",
	}

	// ErrInvalidDBConnection indicates an exception occurred due to an invalid database connection.
	ErrInvalidDBConnection = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "ServiceException",
		Message:  "Data operation exception.",
	}

	// ErrPanicException indicates an abnormal behavior occurred during the operation.
	ErrPanicException = &Err{
		HTTPCode: http.StatusInternalServerError,
		ErrType:  ErrTypeServer,
		Code:     "OperateException",
		Message:  "An abnormal behavior occurred during the operation.",
	}
)

// API Common Errors - these are errors related to API interface usage.
var (
	// ErrUnauthorized indicates that the account is not authorized to perform the action.
	ErrUnauthorized = &Err{
		HTTPCode: http.StatusForbidden,
		ErrType:  ErrTypeSender,
		Code:     "Unauthorized",
		Message:  "Unauthorized account.",
	}

	// ErrPermissionDenied
	ErrPermissionDenied = &Err{
		HTTPCode: http.StatusForbidden,
		ErrType:  ErrTypeSender,
		Code:     "PermissionDenied",
		Message:  "Permission denied.",
	}

	// ErrNoSuchEntity indicates that the requested entity does not exist.
	ErrNoSuchEntity = &Err{
		HTTPCode: http.StatusNotFound,
		ErrType:  ErrTypeSender,
		Code:     "NoSuchEntity",
		Message:  "The request was rejected because it referenced a non-existent `InnerApi`.",
	}

	// ErrInvalidMethod indicates that the HTTP method used is not allowed for the requested resource.
	ErrInvalidMethod = &Err{
		HTTPCode: http.StatusMethodNotAllowed,
		ErrType:  ErrTypeSender,
		Code:     "InvalidMethod",
		Message:  "Method Not Allowed.",
	}

	// ErrInvalidAction indicates that the 'Action' parameter referenced does not exist.
	ErrInvalidAction = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "InvalidAction",
		Message:  "Request was rejected because it referenced an 'Action' that does not exist.",
	}

	// ErrNoSuchVersion indicates that the 'Version' parameter referenced does not exist.
	ErrNoSuchVersion = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "NoSuchVersion",
		Message:  "Request was rejected because it referenced an 'Version' that does not exist.",
	}

	// ErrTokenInvalid indicates that the provided token is invalid.
	ErrTokenInvalid = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "InvalidToken",
		Message:  "Invalid `AccessToken`.",
	}

	// ErrCancelRequest indicates that the user has canceled the request.
	ErrCancelRequest = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "CancelRequest",
		Message:  "User cancel request.",
	}

	// ErrUnsupportedRegion indicates that the region does not provide the corresponding service.
	ErrUnsupportedRegion = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "UnsupportedRegion",
		Message:  "The region does not provide the corresponding service.",
	}

	// ErrNotFoundRegion indicates that the specified region could not be found.
	ErrNotFoundRegion = &Err{
		HTTPCode: http.StatusNotFound,
		ErrType:  ErrTypeSender,
		Code:     "NotFoundRegion",
		Message:  "The region not found.",
	}

	// ErrMissingHeader indicates that a required header is missing from the request.
	ErrMissingHeader = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "MissingHeader",
		Message:  "Request `%s` is missing in Header.",
	}

	// ErrMissingParameter indicates that a required parameter is missing from the request.
	ErrMissingParameter = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "MissingParameter",
		Message:  "Parameter %s must be required.",
	}

	// ErrInvalidParameter indicates that a parameter has an incorrect format or value type.
	ErrInvalidParameter = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "InvalidParameter",
		Message:  "Invalid parameter, incorrect parameter format or value type.",
	}

	// ErrInvalidParameterValue indicates that a parameter has an invalid value.
	ErrInvalidParameterValue = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "InvalidParameterValue",
		Message:  "%s",
	}
)

// Business Errors - these are errors specific to business logic.
var (
	// ErrNotFound indicates that the specified resource could not be found.
	ErrNotFound = &Err{
		HTTPCode: http.StatusNotFound,
		ErrType:  ErrTypeSender,
		Code:     "NotFound", Message: "%s",
	}

	// ErrExceedsLimit indicates that the operation exceeds some defined limit.
	ErrExceedsLimit = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeSender,
		Code:     "ExceedsLimit",
		Message:  "%s",
	}

	//ErrNotFoundResource indicates that the specified resource could not be found.
	ErrNotFoundResource = &Err{
		HTTPCode: http.StatusNotFound,
		ErrType:  ErrTypeSender,
		Code:     "NotFoundResource",
		Message:  "The specified resource %s is not found.",
	}

	// ErrOperateFailed indicates that the operation failed.
	ErrOperateFailed = &Err{
		HTTPCode: http.StatusBadRequest,
		ErrType:  ErrTypeServer,
		Code:     "OperateFailed",
		Message:  "%s",
	}
)
