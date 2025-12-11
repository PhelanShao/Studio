package otel

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	// RequestIDHeader is the header name for request ID.
	RequestIDHeader = "X-Request-ID"

	// UserIDContextKey is the context key for user ID (set by auth middleware).
	UserIDContextKey = "user_id"

	// LabIDContextKey is the context key for lab ID.
	LabIDContextKey = "lab_id"
)

// EnhancedMiddleware returns a Gin middleware that enhances OpenTelemetry spans
// with additional business attributes and records metrics.
func EnhancedMiddleware() gin.HandlerFunc {
	metrics := GetMetrics()

	return func(c *gin.Context) {
		startTime := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			if id, err := uuid.NewV4(); err == nil {
				requestID = id.String()
			}
		}
		c.Header(RequestIDHeader, requestID)

		// Get the current span from context (set by otelgin middleware)
		span := trace.SpanFromContext(c.Request.Context())

		// Add request ID to span
		span.SetAttributes(attribute.String("request.id", requestID))

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Seconds()

		// Get user ID from context (set by auth middleware)
		userID := ""
		if uid, exists := c.Get(UserIDContextKey); exists {
			if uidStr, ok := uid.(string); ok {
				userID = uidStr
			}
		}

		// Add user ID to span if available
		if userID != "" {
			span.SetAttributes(attribute.String("user.id", userID))
		}

		// Get lab ID from context if available
		if labID, exists := c.Get(LabIDContextKey); exists {
			if labIDStr, ok := labID.(string); ok {
				span.SetAttributes(attribute.String("lab.id", labIDStr))
			}
		}

		// Get the matched route pattern (not the actual path with params)
		routePattern := c.FullPath()
		if routePattern == "" {
			routePattern = c.Request.URL.Path
		}

		// Record HTTP metrics
		metrics.RecordHTTPRequest(c.Request.Context(), c.Request.Method, routePattern, c.Writer.Status(), userID)
		metrics.RecordHTTPDuration(c.Request.Context(), c.Request.Method, routePattern, duration)

		// Add response attributes to span
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_content_length", c.Writer.Size()),
		)
	}
}

// SetLabID sets the lab ID in the Gin context for later use in span attributes.
func SetLabID(c *gin.Context, labID string) {
	c.Set(LabIDContextKey, labID)

	// Also add to current span
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(attribute.String("lab.id", labID))
}

// StartSpan starts a new span with the given name and returns the context and span.
// The caller should defer span.End().
func StartSpan(c *gin.Context, spanName string, attrs ...attribute.KeyValue) (trace.Span, func()) {
	tracer := otel.Tracer(MeterName)
	_, span := tracer.Start(c.Request.Context(), spanName, trace.WithAttributes(attrs...))

	return span, func() {
		span.End()
	}
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(c *gin.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request.Context())
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanError records an error on the current span.
func SetSpanError(c *gin.Context, err error) {
	span := trace.SpanFromContext(c.Request.Context())
	span.RecordError(err)
}

