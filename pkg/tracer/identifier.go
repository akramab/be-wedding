package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type requestIDCtx struct{}

// UpdateContextWithRequestID set request ID to given context and return the copy it.
func UpdateContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDCtx{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value(requestIDCtx{}).(string)
	if !ok {
		return ""
	}
	return reqID
}

func TraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasTraceID() {
		return ""
	}
	return spanCtx.TraceID().String()
}

func SpanIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasSpanID() {
		return ""
	}
	return spanCtx.SpanID().String()
}
