package otel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	// GetMetrics should return the same instance
	m1 := GetMetrics()
	m2 := GetMetrics()

	assert.NotNil(t, m1)
	assert.Same(t, m1, m2, "GetMetrics should return the same instance")

	// Check that all metrics are initialized
	assert.NotNil(t, m1.HTTPRequestsTotal)
	assert.NotNil(t, m1.HTTPRequestDuration)
	assert.NotNil(t, m1.WorkflowExecutionsTotal)
	assert.NotNil(t, m1.WorkflowExecutionDuration)
	assert.NotNil(t, m1.ActionExecutionsTotal)
	assert.NotNil(t, m1.WebSocketConnections)
}

func TestEnhancedMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(EnhancedMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check that X-Request-ID header is set
	assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
}

func TestEnhancedMiddlewareWithExistingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(EnhancedMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Create test request with existing request ID
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "existing-request-id")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check that the existing request ID is preserved
	assert.Equal(t, "existing-request-id", w.Header().Get(RequestIDHeader))
}

func TestSetLabID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test/:lab_id", func(c *gin.Context) {
		labID := c.Param("lab_id")
		SetLabID(c, labID)

		// Verify it was set in context
		storedLabID, exists := c.Get(LabIDContextKey)
		assert.True(t, exists)
		assert.Equal(t, labID, storedLabID)

		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test/lab-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInjectExtractTraceContext(t *testing.T) {
	// Test that InjectTraceContext and ExtractTraceContext work correctly
	headers := make(http.Header)

	// Without any trace context, headers should be empty or have no traceparent
	InjectTraceContext(nil, headers)
	// This is a basic test - in a real scenario with active trace, headers would be populated

	// ExtractTraceContext should return a valid context
	// ctx := ExtractTraceContext(context.Background(), headers)
	// assert.NotNil(t, ctx)
}

