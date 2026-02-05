package http

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

func TraceIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			span := trace.SpanFromContext(c.Request().Context())
			if span.SpanContext().HasTraceID() {
				traceID := span.SpanContext().TraceID().String()
				c.Response().Header().Set("X-Trace-ID", traceID)
			}
			return next(c)
		}
	}
}
