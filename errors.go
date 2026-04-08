package intelmesh

import (
	"errors"
	"fmt"
)

// Error codes returned by the IntelMesh API.
const (
	// CodeValidation indicates a validation error.
	CodeValidation = "VALIDATION_ERROR"
	// CodeNotFound indicates the requested resource was not found.
	CodeNotFound = "NOT_FOUND"
	// CodeUnauthorized indicates missing or invalid authentication.
	CodeUnauthorized = "UNAUTHORIZED"
	// CodeForbidden indicates insufficient permissions.
	CodeForbidden = "FORBIDDEN"
	// CodeInternal indicates an internal server error.
	CodeInternal = "INTERNAL"
	// CodeUnavailable indicates the service is temporarily unavailable.
	CodeUnavailable = "UNAVAILABLE"
	// CodeInvalidBody indicates the request body is malformed.
	CodeInvalidBody = "INVALID_BODY"
	// CodeInvalidParam indicates an invalid query or path parameter.
	CodeInvalidParam = "INVALID_PARAM"
)

// IntelMeshError is the base error for all API errors.
type IntelMeshError struct {
	StatusCode int
	Code       string
	Message    string
}

// Error returns a human-readable error string.
func (e *IntelMeshError) Error() string {
	return fmt.Sprintf("intelmesh: %s (HTTP %d): %s", e.Code, e.StatusCode, e.Message)
}

// IsNotFound reports whether the error is a NOT_FOUND error.
func IsNotFound(err error) bool {
	return hasCode(err, CodeNotFound)
}

// IsValidation reports whether the error is a validation error.
func IsValidation(err error) bool {
	return hasCode(err, CodeValidation) ||
		hasCode(err, CodeInvalidBody) ||
		hasCode(err, CodeInvalidParam)
}

// IsUnauthorized reports whether the error is an UNAUTHORIZED error.
func IsUnauthorized(err error) bool {
	return hasCode(err, CodeUnauthorized)
}

// IsForbidden reports whether the error is a FORBIDDEN error.
func IsForbidden(err error) bool {
	return hasCode(err, CodeForbidden)
}

// IsUnavailable reports whether the error is an UNAVAILABLE error.
func IsUnavailable(err error) bool {
	return hasCode(err, CodeUnavailable)
}

// IsInternal reports whether the error is an INTERNAL error.
func IsInternal(err error) bool {
	return hasCode(err, CodeInternal)
}

func hasCode(err error, code string) bool {
	var e *IntelMeshError
	if errors.As(err, &e) {
		return e.Code == code
	}

	return false
}
