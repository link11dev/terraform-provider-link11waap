package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- APIError ---

func TestAPIError_Error(t *testing.T) {
	e := &APIError{Code: 400, Message: "bad request"}
	assert.Equal(t, "API error (code 400): bad request", e.Error())
}

// --- ValidationError ---

func TestValidationError_Error_WithDetail(t *testing.T) {
	e := &ValidationError{
		Detail: []ValidationErrorDetail{
			{Msg: "field required", Type: "value_error"},
		},
	}
	assert.Equal(t, "validation error: field required", e.Error())
}

func TestValidationError_Error_Empty(t *testing.T) {
	e := &ValidationError{}
	assert.Equal(t, "validation error", e.Error())
}

// --- DetailedError ---

func TestDetailedError_Error(t *testing.T) {
	e := &DetailedError{Message: "publish failed", Detail: "timeout"}
	assert.Equal(t, "publish failed: timeout", e.Error())
}

// --- ParseErrorResponse ---

func TestParseErrorResponse_400(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"code":400,"message":"invalid input"}`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.Code)
	assert.Equal(t, "invalid input", apiErr.Message)
}

func TestParseErrorResponse_400_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`not json`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad response")
}

func TestParseErrorResponse_404(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"message":"not found"}`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.Code)
}

func TestParseErrorResponse_404_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`plain text not found`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.Code)
	assert.Equal(t, "plain text not found", apiErr.Message)
}

func TestParseErrorResponse_422(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       io.NopCloser(strings.NewReader(`{"detail":[{"loc":["body","name"],"msg":"field required","type":"value_error"}]}`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	valErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Contains(t, valErr.Error(), "field required")
}

func TestParseErrorResponse_422_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       io.NopCloser(strings.NewReader(`not json`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestParseErrorResponse_500(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(`{"code":500,"message":"internal error"}`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 500, apiErr.Code)
}

func TestParseErrorResponse_500_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(`crash`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestParseErrorResponse_503(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(strings.NewReader(`{"message":"publish in progress","detail":"busy"}`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	detErr, ok := err.(*DetailedError)
	require.True(t, ok)
	assert.Contains(t, detErr.Error(), "publish in progress")
}

func TestParseErrorResponse_503_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(strings.NewReader(`not json`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestParseErrorResponse_UnexpectedStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader(`forbidden`)),
	}
	err := ParseErrorResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

// --- IsNotFoundError ---

func TestIsNotFoundError_True(t *testing.T) {
	err := &APIError{Code: 404, Message: "not found"}
	assert.True(t, IsNotFoundError(err))
}

func TestIsNotFoundError_False_DifferentCode(t *testing.T) {
	err := &APIError{Code: 400, Message: "bad request"}
	assert.False(t, IsNotFoundError(err))
}

func TestIsNotFoundError_False_DifferentType(t *testing.T) {
	err := &ValidationError{}
	assert.False(t, IsNotFoundError(err))
}
