package tracing

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

/*
When a service receives a request:

Checks if the X-Trace-Id is pasted on the envelope (HTTP header)
If not (in the case where the API gateway is the first to receive the request from the user) -> Generates a new TraceID
If already present (passed over by another service) -> Uses that TraceID
Takes the X-Span-Id of the calling service as its ParentSpanID, and simultaneously generates a new SpanID for itself
then include all three pieces of information in the context
so that the functions inside can easily retrieve and use them
*/
func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		traceIDStr := c.Get(HeaderTraceID)
		var traceID uuid.UUID
		if traceIDStr == "" {
			traceID = uuid.New()
		} else {
			parsed, err := uuid.Parse(traceIDStr)
			// the X-Trace-Id header exists (it's not empty), but the value inside can be an invalid UUID
			// handdle if traceIDStr is invalid
			if err != nil {
				traceID = uuid.New()
			} else {
				traceID = parsed
			}
		}

		parentSpanIDStr := c.Get(HeaderSpanID)
		spanID := uuid.New()

		ctx := context.WithValue(c.Context(), traceIDKey, traceID)
		ctx = context.WithValue(ctx, spanIDKey, spanID)
		if parentSpanIDStr != "" {
			ctx = context.WithValue(ctx, parentSpanIDKey, parentSpanIDStr)
		}
		c.SetContext(ctx)

		return c.Next()
	}
}
