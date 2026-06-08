// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"errors"
	"net/http"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
	"github.com/MustardSeedNetworks/stem/internal/services/orchestrator/dataplane"
)

// ErrorCode represents a machine-readable error identifier.
// Clients can use these codes to programmatically handle specific error conditions.
type ErrorCode string

// Standard error codes for API responses.
const (
	// ErrCodeAuthFailed indicates authentication failed.
	ErrCodeAuthFailed ErrorCode = "AUTH_FAILED"
	// ErrCodeAuthExpired indicates the authentication token has expired.
	ErrCodeAuthExpired ErrorCode = "AUTH_EXPIRED"
	// ErrCodePermissionDenied indicates the user lacks required permissions.
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// ErrCodeInvalidRequest indicates the request was invalid.
	ErrCodeInvalidRequest ErrorCode = "INVALID_REQUEST"
	// ErrCodeNotFound indicates the requested resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeMethodNotAllowed indicates the HTTP method is not allowed.
	ErrCodeMethodNotAllowed ErrorCode = "METHOD_NOT_ALLOWED"
	// ErrCodeConflict indicates a resource conflict.
	ErrCodeConflict ErrorCode = "CONFLICT"

	// ErrCodeRateLimited indicates the rate limit was exceeded.
	ErrCodeRateLimited ErrorCode = "RATE_LIMITED"

	// ErrCodeInternalError indicates an internal server error.
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	// ErrCodeServiceUnavailable indicates the service is temporarily unavailable.
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// HTTPErrorResponse represents a standardized error response for the API.
// All error responses from the API follow this format to ensure consistency.
type HTTPErrorResponse struct {
	// Error is a short, human-readable error type (e.g., "Unauthorized", "Bad Request").
	Error string `json:"error"`
	// Code is a machine-readable error code for client-side handling.
	Code ErrorCode `json:"code"`
	// Message is a human-readable description of the error.
	// This message is sanitized and safe to display to end users.
	Message string `json:"message"`
}

// Error represents an error with associated HTTP status and error code.
// It implements the error interface and can be used throughout the application.
type Error struct {
	// HTTPStatus is the HTTP status code to return.
	HTTPStatus int
	// Code is the machine-readable error code.
	Code ErrorCode
	// Message is the user-facing error message.
	Message string
	// InternalErr is the underlying error (not exposed to clients).
	InternalErr error
}

// NewError creates a new Error with the given parameters.
func NewError(status int, code ErrorCode, message string) *Error {
	return &Error{
		HTTPStatus:  status,
		Code:        code,
		Message:     message,
		InternalErr: nil,
	}
}

// NewErrorWithCause creates a new Error that wraps an underlying error.
// The internal error is logged but not exposed to clients.
func NewErrorWithCause(status int, code ErrorCode, message string, cause error) *Error {
	return &Error{
		HTTPStatus:  status,
		Code:        code,
		Message:     message,
		InternalErr: cause,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.InternalErr != nil {
		return e.InternalErr.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error for [errors.Is]/As support.
func (e *Error) Unwrap() error {
	return e.InternalErr
}

// ToResponse converts the Error to an HTTPErrorResponse suitable for JSON encoding.
func (e *Error) ToResponse() HTTPErrorResponse {
	return HTTPErrorResponse{
		Error:   http.StatusText(e.HTTPStatus),
		Code:    e.Code,
		Message: e.Message,
	}
}

// Predefined API errors for common cases.
var (
	// ErrAuthFailed is returned when authentication fails.
	ErrAuthFailed = &Error{
		HTTPStatus:  http.StatusUnauthorized,
		Code:        ErrCodeAuthFailed,
		Message:     "Authentication failed",
		InternalErr: nil,
	}

	// ErrAuthExpired is returned when the authentication token has expired.
	ErrAuthExpired = &Error{
		HTTPStatus:  http.StatusUnauthorized,
		Code:        ErrCodeAuthExpired,
		Message:     "Authentication token has expired",
		InternalErr: nil,
	}

	// ErrAuthRevoked is returned when the authentication token has been revoked.
	ErrAuthRevoked = &Error{
		HTTPStatus:  http.StatusUnauthorized,
		Code:        ErrCodeAuthFailed,
		Message:     "Authentication token has been revoked",
		InternalErr: nil,
	}

	// ErrPermissionDenied is returned when the user lacks required permissions.
	ErrPermissionDenied = &Error{
		HTTPStatus:  http.StatusForbidden,
		Code:        ErrCodePermissionDenied,
		Message:     "Insufficient permissions",
		InternalErr: nil,
	}

	// ErrMethodNotAllowed is returned when the HTTP method is not allowed.
	ErrMethodNotAllowed = &Error{
		HTTPStatus:  http.StatusMethodNotAllowed,
		Code:        ErrCodeMethodNotAllowed,
		Message:     "Method not allowed",
		InternalErr: nil,
	}

	// ErrInvalidJSON is returned when the request body contains invalid JSON.
	ErrInvalidJSON = &Error{
		HTTPStatus:  http.StatusBadRequest,
		Code:        ErrCodeInvalidRequest,
		Message:     "Invalid JSON in request body",
		InternalErr: nil,
	}

	// ErrRequestTooLarge is returned when the request body is too large.
	ErrRequestTooLarge = &Error{
		HTTPStatus:  http.StatusRequestEntityTooLarge,
		Code:        ErrCodeInvalidRequest,
		Message:     "Request body too large",
		InternalErr: nil,
	}

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = &Error{
		HTTPStatus:  http.StatusNotFound,
		Code:        ErrCodeNotFound,
		Message:     "Resource not found",
		InternalErr: nil,
	}

	// ErrConflict is returned when there is a resource conflict.
	ErrConflict = &Error{
		HTTPStatus:  http.StatusConflict,
		Code:        ErrCodeConflict,
		Message:     "Resource conflict",
		InternalErr: nil,
	}

	// ErrRateLimited is returned when the rate limit is exceeded.
	ErrRateLimited = &Error{
		HTTPStatus:  http.StatusTooManyRequests,
		Code:        ErrCodeRateLimited,
		Message:     "Too many requests",
		InternalErr: nil,
	}

	// ErrInternalError is returned when an internal server error occurs.
	ErrInternalError = &Error{
		HTTPStatus:  http.StatusInternalServerError,
		Code:        ErrCodeInternalError,
		Message:     "An internal error occurred",
		InternalErr: nil,
	}

	// ErrServiceUnavailable is returned when the service is temporarily unavailable.
	ErrServiceUnavailable = &Error{
		HTTPStatus:  http.StatusServiceUnavailable,
		Code:        ErrCodeServiceUnavailable,
		Message:     "Service temporarily unavailable",
		InternalErr: nil,
	}
)

// InvalidRequestError creates a bad request error with a custom message.
func InvalidRequestError(message string) *Error {
	return NewError(http.StatusBadRequest, ErrCodeInvalidRequest, message)
}

// NotFoundError creates a not found error with a custom message.
func NotFoundError(message string) *Error {
	return NewError(http.StatusNotFound, ErrCodeNotFound, message)
}

// ConflictError creates a conflict error with a custom message.
func ConflictError(message string) *Error {
	return NewError(http.StatusConflict, ErrCodeConflict, message)
}

// InternalError creates an internal error that wraps the underlying cause.
// The cause is logged but a generic message is returned to clients.
func InternalError(cause error) *Error {
	return &Error{
		HTTPStatus:  http.StatusInternalServerError,
		Code:        ErrCodeInternalError,
		Message:     "An internal error occurred",
		InternalErr: cause,
	}
}

// MapAuthError converts an auth package error to an appropriate Error.
// This centralizes the mapping of authentication errors to HTTP responses.
func MapAuthError(err error) *Error {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Invalid username or password",
			InternalErr: err,
		}
	case errors.Is(err, auth.ErrInvalidToken):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Invalid authentication token",
			InternalErr: err,
		}
	case errors.Is(err, auth.ErrTokenExpired):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthExpired,
			Message:     "Authentication token has expired",
			InternalErr: err,
		}
	case errors.Is(err, auth.ErrTokenRevoked):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Authentication token has been revoked",
			InternalErr: err,
		}
	case errors.Is(err, errMissingAuthToken):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Missing authentication token",
			InternalErr: err,
		}
	case errors.Is(err, errInvalidAuthHeader):
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Invalid authorization header format",
			InternalErr: err,
		}
	default:
		// Generic auth error - don't leak internal details.
		return &Error{
			HTTPStatus:  http.StatusUnauthorized,
			Code:        ErrCodeAuthFailed,
			Message:     "Authentication failed",
			InternalErr: err,
		}
	}
}

// MapTestError converts a test execution error to an appropriate Error.
// This centralizes the mapping of test-related errors to HTTP responses.
func MapTestError(err error) *Error {
	switch {
	case errors.Is(err, errTestAlreadyRunning):
		return &Error{
			HTTPStatus:  http.StatusConflict,
			Code:        ErrCodeConflict,
			Message:     "A test is already running",
			InternalErr: err,
		}
	case errors.Is(err, dataplane.ErrNotSupported):
		return &Error{
			HTTPStatus:  http.StatusServiceUnavailable,
			Code:        ErrCodeServiceUnavailable,
			Message:     "Test execution requires Linux with CGO support",
			InternalErr: err,
		}
	default:
		// Generic test error - don't leak internal details.
		return &Error{
			HTTPStatus:  http.StatusInternalServerError,
			Code:        ErrCodeInternalError,
			Message:     "Failed to execute test",
			InternalErr: err,
		}
	}
}

// WriteError writes an Error to the response as JSON.
// It logs internal errors for debugging while returning sanitized messages to clients.
func WriteError(w http.ResponseWriter, apiErr *Error) {
	// Log internal error details for debugging (not exposed to client).
	if apiErr.InternalErr != nil {
		logging.Error("API error",
			"status", apiErr.HTTPStatus,
			"code", apiErr.Code,
			"message", apiErr.Message,
			"internal_error", apiErr.InternalErr,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	writeJSON(w, apiErr.ToResponse())
}

// HandleError is a convenience function that writes an error response.
// It attempts to convert the error to an Error; if not possible,
// it treats it as an internal error.
func HandleError(w http.ResponseWriter, err error) {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		WriteError(w, apiErr)
		return
	}

	// Wrap unknown errors as internal errors.
	WriteError(w, InternalError(err))
}

// WriteMethodNotAllowed writes a 405 Method Not Allowed response.
func WriteMethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, ErrMethodNotAllowed)
}

// WriteNotFound writes a 404 Not Found response with a custom message.
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, NotFoundError(message))
}

// WriteInvalidRequest writes a 400 Bad Request response with a custom message.
func WriteInvalidRequest(w http.ResponseWriter, message string) {
	WriteError(w, InvalidRequestError(message))
}

// WriteConflict writes a 409 Conflict response with a custom message.
func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, ConflictError(message))
}

// WriteInternalError writes a 500 Internal Server Error response.
// The actual error is logged but not exposed to the client.
func WriteInternalError(w http.ResponseWriter, cause error) {
	WriteError(w, InternalError(cause))
}

// WriteAuthError writes an authentication error response.
// It maps the auth error to the appropriate API error.
func WriteAuthError(w http.ResponseWriter, err error) {
	WriteError(w, MapAuthError(err))
}
