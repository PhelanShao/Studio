package otel

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// PropagatingHTTPClient returns a resty client that propagates trace context.
func PropagatingHTTPClient() *resty.Client {
	client := resty.New()

	// Add trace context propagation middleware
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		ctx := req.Context()
		if ctx == nil {
			return nil
		}

		// Inject trace context into request headers
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

		return nil
	})

	// Add response tracing middleware
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		ctx := resp.Request.Context()
		if ctx == nil {
			return nil
		}

		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.Int("http.status_code", resp.StatusCode()),
			attribute.String("http.url", resp.Request.URL),
		)

		return nil
	})

	return client
}

// InjectTraceContext injects the trace context from ctx into the HTTP headers.
func InjectTraceContext(ctx context.Context, headers http.Header) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractTraceContext extracts the trace context from HTTP headers into a context.
func ExtractTraceContext(ctx context.Context, headers http.Header) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// StartClientSpan starts a new span for an outgoing HTTP request.
// Returns the updated context and a cleanup function that should be deferred.
func StartClientSpan(ctx context.Context, method, url string) (context.Context, trace.Span, func()) {
	tracer := otel.Tracer(MeterName)

	ctx, span := tracer.Start(ctx, "HTTP "+method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
		),
	)

	return ctx, span, func() {
		span.End()
	}
}

// SetClientSpanStatus sets the status of a client span based on the HTTP response.
func SetClientSpanStatus(span trace.Span, statusCode int, err error) {
	if err != nil {
		span.RecordError(err)
	}
	span.SetAttributes(attribute.Int("http.status_code", statusCode))
}

// RestyClientWithTracing returns a configured resty client with full tracing support.
// This includes trace context propagation and span creation for each request.
func RestyClientWithTracing() *resty.Client {
	client := resty.New()

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		ctx := req.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// Start a new span for the outgoing request
		tracer := otel.Tracer(MeterName)
		ctx, span := tracer.Start(ctx, "HTTP "+req.Method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL),
			),
		)

		// Store span in request context for later access
		req.SetContext(ctx)

		// Inject trace context into request headers
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

		// Store span reference for cleanup in OnAfterResponse
		c.SetCloseConnection(false)
		req.SetContext(context.WithValue(ctx, "otel_span", span))

		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		ctx := resp.Request.Context()
		if ctx == nil {
			return nil
		}

		// Retrieve and end the span
		if spanVal := ctx.Value("otel_span"); spanVal != nil {
			if span, ok := spanVal.(trace.Span); ok {
				span.SetAttributes(
					attribute.Int("http.status_code", resp.StatusCode()),
					attribute.Int64("http.response_content_length", resp.Size()),
				)
				span.End()
			}
		}

		return nil
	})

	return client
}

