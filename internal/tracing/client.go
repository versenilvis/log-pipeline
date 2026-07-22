package tracing

import (
	"context"
	"net/http"
)

/*
When Service A wants to make an HTTP call to Service B (for example, order-service calling payment-service):

Retrieve A's current TraceID and SpanID from the context
Add these two codes to the X-Trace-Id and X-Span-Id headers of the request to be sent to B
Therefore, when B receives the email, B will immediately know that A's SpanID is B's ParentSpanID
*/
func InjectHeaders(ctx context.Context, req *http.Request) {
	traceID := TraceIDFromContext(ctx)
	spanID := SpanIDFromContext(ctx)

	req.Header.Set(HeaderTraceID, traceID.String())
	req.Header.Set(HeaderSpanID, spanID.String())
}
