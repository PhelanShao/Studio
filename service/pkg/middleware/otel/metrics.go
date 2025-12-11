// Package otel provides enhanced OpenTelemetry instrumentation for Studio.
package otel

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	// MeterName is the name of the meter used for Studio metrics.
	MeterName = "github.com/scienceol/studio/service"
)

// Metrics holds all the business metrics for Studio.
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     metric.Int64Counter
	HTTPRequestDuration   metric.Float64Histogram

	// Workflow metrics
	WorkflowExecutionsTotal    metric.Int64Counter
	WorkflowExecutionDuration  metric.Float64Histogram

	// Action metrics
	ActionExecutionsTotal metric.Int64Counter

	// WebSocket metrics
	WebSocketConnections metric.Int64UpDownCounter
}

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// GetMetrics returns the global Metrics instance.
// It initializes the metrics on first call.
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = initMetrics()
	})
	return globalMetrics
}

func initMetrics() *Metrics {
	meter := otel.Meter(MeterName)
	m := &Metrics{}

	var err error

	// HTTP metrics
	m.HTTPRequestsTotal, err = meter.Int64Counter(
		"studio_http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		otel.Handle(err)
	}

	m.HTTPRequestDuration, err = meter.Float64Histogram(
		"studio_http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		otel.Handle(err)
	}

	// Workflow metrics
	m.WorkflowExecutionsTotal, err = meter.Int64Counter(
		"studio_workflow_executions_total",
		metric.WithDescription("Total number of workflow executions"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		otel.Handle(err)
	}

	m.WorkflowExecutionDuration, err = meter.Float64Histogram(
		"studio_workflow_execution_duration_seconds",
		metric.WithDescription("Workflow execution duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300),
	)
	if err != nil {
		otel.Handle(err)
	}

	// Action metrics
	m.ActionExecutionsTotal, err = meter.Int64Counter(
		"studio_action_executions_total",
		metric.WithDescription("Total number of action executions"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		otel.Handle(err)
	}

	// WebSocket metrics
	m.WebSocketConnections, err = meter.Int64UpDownCounter(
		"studio_websocket_connections",
		metric.WithDescription("Current number of WebSocket connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		otel.Handle(err)
	}

	return m
}

// RecordHTTPRequest records an HTTP request metric.
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, path string, statusCode int, userID string) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.Int("http.status_code", statusCode),
	}
	if userID != "" {
		attrs = append(attrs, attribute.String("user.id", userID))
	}
	m.HTTPRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordHTTPDuration records HTTP request duration.
func (m *Metrics) RecordHTTPDuration(ctx context.Context, method, path string, durationSeconds float64) {
	m.HTTPRequestDuration.Record(ctx, durationSeconds, metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", path),
	))
}

// RecordWorkflowExecution records a workflow execution metric.
func (m *Metrics) RecordWorkflowExecution(ctx context.Context, labID, status string) {
	m.WorkflowExecutionsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("lab.id", labID),
		attribute.String("status", status),
	))
}

// RecordWorkflowDuration records workflow execution duration.
func (m *Metrics) RecordWorkflowDuration(ctx context.Context, labID string, durationSeconds float64) {
	m.WorkflowExecutionDuration.Record(ctx, durationSeconds, metric.WithAttributes(
		attribute.String("lab.id", labID),
	))
}

// RecordActionExecution records an action execution metric.
func (m *Metrics) RecordActionExecution(ctx context.Context, actionType, status string) {
	m.ActionExecutionsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("action.type", actionType),
		attribute.String("status", status),
	))
}

// WebSocketConnected increments the WebSocket connection counter.
func (m *Metrics) WebSocketConnected(ctx context.Context, connType string) {
	m.WebSocketConnections.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", connType),
	))
}

// WebSocketDisconnected decrements the WebSocket connection counter.
func (m *Metrics) WebSocketDisconnected(ctx context.Context, connType string) {
	m.WebSocketConnections.Add(ctx, -1, metric.WithAttributes(
		attribute.String("type", connType),
	))
}

