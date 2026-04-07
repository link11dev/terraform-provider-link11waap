package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error response from the API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (code %d): %s", e.Code, e.Message)
}

// ValidationError represents a validation error from the API
type ValidationError struct {
	Detail []ValidationErrorDetail `json:"detail"`
}

// ValidationErrorDetail represents a single validation error detail
type ValidationErrorDetail struct {
	Loc  []interface{} `json:"loc"`
	Msg  string        `json:"msg"`
	Type string        `json:"type"`
}

func (e *ValidationError) Error() string {
	if len(e.Detail) == 0 {
		return "validation error"
	}
	return fmt.Sprintf("validation error: %s", e.Detail[0].Msg)
}

// DetailedError represents a detailed error response (used for 503 publish errors)
type DetailedError struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Data    struct {
		Message string `json:"message"`
	} `json:"data"`
}

func (e *DetailedError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Detail)
}

// ParseErrorResponse parses an error response from the API
func ParseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading error response: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("bad response: %s", string(body))
		}
		return &apiErr

	case http.StatusNotFound:
		var apiErr APIError
		apiErr.Code = http.StatusNotFound
		if err := json.Unmarshal(body, &apiErr); err != nil {
			apiErr.Message = string(body)
		}
		return &apiErr

	case http.StatusUnprocessableEntity:
		var validationErr ValidationError
		if err := json.Unmarshal(body, &validationErr); err != nil {
			return fmt.Errorf("validation error: %s", string(body))
		}
		return &validationErr

	case http.StatusInternalServerError:
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("internal server error: %s", string(body))
		}
		return &apiErr

	case http.StatusServiceUnavailable:
		var detailedErr DetailedError
		if err := json.Unmarshal(body, &detailedErr); err != nil {
			return fmt.Errorf("service unavailable: %s", string(body))
		}
		return &detailedErr

	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}

// IsNotFoundError checks if the error indicates a resource was not found
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == 404
	}
	return false
}
