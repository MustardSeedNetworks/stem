// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krisarmstrong/stem/internal/auth"
	"github.com/krisarmstrong/stem/internal/services/orchestrator/dataplane"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		apiErr   *Error
		expected string
	}{
		{
			name: "with internal error",
			apiErr: &Error{
				HTTPStatus:  http.StatusBadRequest,
				Code:        ErrCodeInvalidRequest,
				Message:     "User message",
				InternalErr: errors.New("internal details"),
			},
			expected: "internal details",
		},
		{
			name: "without internal error",
			apiErr: &Error{
				HTTPStatus:  http.StatusBadRequest,
				Code:        ErrCodeInvalidRequest,
				Message:     "User message",
				InternalErr: nil,
			},
			expected: "User message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.apiErr.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	innerErr := errors.New("inner error")
	apiErr := &Error{
		HTTPStatus:  http.StatusInternalServerError,
		Code:        ErrCodeInternalError,
		Message:     "An error occurred",
		InternalErr: innerErr,
	}

	unwrapped := apiErr.Unwrap()
	if !errors.Is(unwrapped, innerErr) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}

	// Test errors.Is support.
	if !errors.Is(apiErr, innerErr) {
		t.Error("errors.Is should return true for wrapped error")
	}
}

func TestError_ToResponse(t *testing.T) {
	t.Parallel()

	apiErr := &Error{
		HTTPStatus:  http.StatusUnauthorized,
		Code:        ErrCodeAuthFailed,
		Message:     "Authentication failed",
		InternalErr: errors.New("secret internal details"),
	}

	resp := apiErr.ToResponse()

	if resp.Error != "Unauthorized" {
		t.Errorf("Error = %q, want %q", resp.Error, "Unauthorized")
	}
	if resp.Code != ErrCodeAuthFailed {
		t.Errorf("Code = %q, want %q", resp.Code, ErrCodeAuthFailed)
	}
	if resp.Message != "Authentication failed" {
		t.Errorf("Message = %q, want %q", resp.Message, "Authentication failed")
	}
}

func TestNewError(t *testing.T) {
	t.Parallel()

	apiErr := NewError(http.StatusNotFound, ErrCodeNotFound, "Resource not found")

	if apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, http.StatusNotFound)
	}
	if apiErr.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", apiErr.Code, ErrCodeNotFound)
	}
	if apiErr.Message != "Resource not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Resource not found")
	}
	if apiErr.InternalErr != nil {
		t.Error("InternalErr should be nil")
	}
}

func TestNewErrorWithCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("database connection failed")
	apiErr := NewErrorWithCause(
		http.StatusInternalServerError,
		ErrCodeInternalError,
		"An error occurred",
		cause,
	)

	if !errors.Is(apiErr.InternalErr, cause) {
		t.Error("InternalErr should match cause")
	}
	if !errors.Is(apiErr, cause) {
		t.Error("errors.Is should return true for cause")
	}
}

func TestErrorHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		createErr    func() *Error
		expectedCode ErrorCode
		expectedHTTP int
	}{
		{
			name:         "InvalidRequestError",
			createErr:    func() *Error { return InvalidRequestError("bad input") },
			expectedCode: ErrCodeInvalidRequest,
			expectedHTTP: http.StatusBadRequest,
		},
		{
			name:         "NotFoundError",
			createErr:    func() *Error { return NotFoundError("item not found") },
			expectedCode: ErrCodeNotFound,
			expectedHTTP: http.StatusNotFound,
		},
		{
			name:         "ConflictError",
			createErr:    func() *Error { return ConflictError("already exists") },
			expectedCode: ErrCodeConflict,
			expectedHTTP: http.StatusConflict,
		},
		{
			name: "InternalError",
			createErr: func() *Error {
				return InternalError(errors.New("db error"))
			},
			expectedCode: ErrCodeInternalError,
			expectedHTTP: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiErr := tt.createErr()

			if apiErr.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.expectedCode)
			}
			if apiErr.HTTPStatus != tt.expectedHTTP {
				t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, tt.expectedHTTP)
			}
		})
	}
}

func TestMapAuthError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		expectedCode ErrorCode
		expectedHTTP int
	}{
		{
			name:         "invalid credentials",
			err:          auth.ErrInvalidCredentials,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "invalid token",
			err:          auth.ErrInvalidToken,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "token expired",
			err:          auth.ErrTokenExpired,
			expectedCode: ErrCodeAuthExpired,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "token revoked",
			err:          auth.ErrTokenRevoked,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "missing auth token",
			err:          errMissingAuthToken,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "invalid auth header",
			err:          errInvalidAuthHeader,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "unknown auth error",
			err:          errors.New("some unknown error"),
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiErr := MapAuthError(tt.err)

			if apiErr.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.expectedCode)
			}
			if apiErr.HTTPStatus != tt.expectedHTTP {
				t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, tt.expectedHTTP)
			}
			// Verify internal error is preserved.
			if !errors.Is(apiErr, tt.err) {
				t.Error("Internal error should be preserved")
			}
		})
	}
}

func TestMapTestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		expectedCode ErrorCode
		expectedHTTP int
	}{
		{
			name:         "test already running",
			err:          errTestAlreadyRunning,
			expectedCode: ErrCodeConflict,
			expectedHTTP: http.StatusConflict,
		},
		{
			name:         "not supported",
			err:          dataplane.ErrNotSupported,
			expectedCode: ErrCodeServiceUnavailable,
			expectedHTTP: http.StatusServiceUnavailable,
		},
		{
			name:         "unknown test error",
			err:          errors.New("some test error"),
			expectedCode: ErrCodeInternalError,
			expectedHTTP: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiErr := MapTestError(tt.err)

			if apiErr.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.expectedCode)
			}
			if apiErr.HTTPStatus != tt.expectedHTTP {
				t.Errorf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, tt.expectedHTTP)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	apiErr := &Error{
		HTTPStatus:  http.StatusBadRequest,
		Code:        ErrCodeInvalidRequest,
		Message:     "Invalid input",
		InternalErr: errors.New("internal details that should not be exposed"),
	}

	rec := httptest.NewRecorder()
	WriteError(rec, apiErr)

	// Check status code.
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	// Check content type.
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	// Check response body.
	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != "Bad Request" {
		t.Errorf("Error = %q, want %q", resp.Error, "Bad Request")
	}
	if resp.Code != ErrCodeInvalidRequest {
		t.Errorf("Code = %q, want %q", resp.Code, ErrCodeInvalidRequest)
	}
	if resp.Message != "Invalid input" {
		t.Errorf("Message = %q, want %q", resp.Message, "Invalid input")
	}

	// Verify internal error is NOT exposed in response.
	body := rec.Body.String()
	if contains := "internal details"; containsString(body, contains) {
		t.Error("Response should not contain internal error details")
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	t.Run("with Error", func(t *testing.T) {
		t.Parallel()
		apiErr := NotFoundError("item not found")
		rec := httptest.NewRecorder()
		HandleError(rec, apiErr)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		t.Parallel()
		genericErr := errors.New("some internal error")
		rec := httptest.NewRecorder()
		HandleError(rec, genericErr)

		// Should be wrapped as internal error.
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}

		var resp HTTPErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != ErrCodeInternalError {
			t.Errorf("Code = %q, want %q", resp.Code, ErrCodeInternalError)
		}
		// Verify error details are not leaked.
		if resp.Message == "some internal error" {
			t.Error("Internal error message should not be exposed")
		}
	})
}

func TestWriteMethodNotAllowed(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteMethodNotAllowed(rec)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeMethodNotAllowed {
		t.Errorf("Code = %q, want %q", resp.Code, ErrCodeMethodNotAllowed)
	}
}

func TestWriteNotFound(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteNotFound(rec, "User not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Message != "User not found" {
		t.Errorf("Message = %q, want %q", resp.Message, "User not found")
	}
}

func TestWriteInvalidRequest(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteInvalidRequest(rec, "Invalid email format")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Message != "Invalid email format" {
		t.Errorf("Message = %q, want %q", resp.Message, "Invalid email format")
	}
}

func TestWriteConflict(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteConflict(rec, "Resource already exists")

	if rec.Code != http.StatusConflict {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusConflict)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Message != "Resource already exists" {
		t.Errorf("Message = %q, want %q", resp.Message, "Resource already exists")
	}
}

func TestWriteInternalError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteInternalError(rec, errors.New("database connection failed"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify error details are not leaked.
	if resp.Message == "database connection failed" {
		t.Error("Internal error message should not be exposed")
	}
	if resp.Message != "An internal error occurred" {
		t.Errorf("Message = %q, want %q", resp.Message, "An internal error occurred")
	}
}

func TestWriteAuthError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteAuthError(rec, auth.ErrTokenExpired)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp HTTPErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeAuthExpired {
		t.Errorf("Code = %q, want %q", resp.Code, ErrCodeAuthExpired)
	}
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		apiErr       *Error
		expectedCode ErrorCode
		expectedHTTP int
	}{
		{
			name:         "ErrAuthFailed",
			apiErr:       ErrAuthFailed,
			expectedCode: ErrCodeAuthFailed,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "ErrAuthExpired",
			apiErr:       ErrAuthExpired,
			expectedCode: ErrCodeAuthExpired,
			expectedHTTP: http.StatusUnauthorized,
		},
		{
			name:         "ErrPermissionDenied",
			apiErr:       ErrPermissionDenied,
			expectedCode: ErrCodePermissionDenied,
			expectedHTTP: http.StatusForbidden,
		},
		{
			name:         "ErrMethodNotAllowed",
			apiErr:       ErrMethodNotAllowed,
			expectedCode: ErrCodeMethodNotAllowed,
			expectedHTTP: http.StatusMethodNotAllowed,
		},
		{
			name:         "ErrInvalidJSON",
			apiErr:       ErrInvalidJSON,
			expectedCode: ErrCodeInvalidRequest,
			expectedHTTP: http.StatusBadRequest,
		},
		{
			name:         "ErrRequestTooLarge",
			apiErr:       ErrRequestTooLarge,
			expectedCode: ErrCodeInvalidRequest,
			expectedHTTP: http.StatusRequestEntityTooLarge,
		},
		{
			name:         "ErrNotFound",
			apiErr:       ErrNotFound,
			expectedCode: ErrCodeNotFound,
			expectedHTTP: http.StatusNotFound,
		},
		{
			name:         "ErrConflict",
			apiErr:       ErrConflict,
			expectedCode: ErrCodeConflict,
			expectedHTTP: http.StatusConflict,
		},
		{
			name:         "ErrRateLimited",
			apiErr:       ErrRateLimited,
			expectedCode: ErrCodeRateLimited,
			expectedHTTP: http.StatusTooManyRequests,
		},
		{
			name:         "ErrInternalError",
			apiErr:       ErrInternalError,
			expectedCode: ErrCodeInternalError,
			expectedHTTP: http.StatusInternalServerError,
		},
		{
			name:         "ErrServiceUnavailable",
			apiErr:       ErrServiceUnavailable,
			expectedCode: ErrCodeServiceUnavailable,
			expectedHTTP: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.apiErr.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", tt.apiErr.Code, tt.expectedCode)
			}
			if tt.apiErr.HTTPStatus != tt.expectedHTTP {
				t.Errorf("HTTPStatus = %d, want %d", tt.apiErr.HTTPStatus, tt.expectedHTTP)
			}
		})
	}
}

func TestErrorSanitization(t *testing.T) {
	t.Parallel()

	// Test that internal details are never exposed in responses.
	sensitiveErrors := []error{
		errors.New("SELECT id, name, email FROM users WHERE id = 1"),
		errors.New("connection refused: 10.0.0.1:5432"),
		errors.New("file not found: /etc/passwd"),
		errors.New("panic: runtime error: invalid memory address"),
	}

	for _, sensitiveErr := range sensitiveErrors {
		rec := httptest.NewRecorder()
		WriteInternalError(rec, sensitiveErr)

		body := rec.Body.String()

		// Verify none of the sensitive content appears in response.
		if containsString(body, sensitiveErr.Error()) {
			t.Errorf("Response leaked sensitive error: %s", sensitiveErr.Error())
		}

		// Verify we get a generic message.
		var resp HTTPErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Message != "An internal error occurred" {
			t.Errorf("Message = %q, want generic message", resp.Message)
		}
	}
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
