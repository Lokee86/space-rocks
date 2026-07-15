package operationcontext

import (
	"context"
	"testing"
)

func TestWithAndFromStoreImmutableOperationValues(t *testing.T) {
	want := Values{TraceID: "trace", RequestID: "request", Route: "/route"}
	ctx := With(context.Background(), want)
	want.TraceID = "changed"

	got, ok := From(ctx)
	if !ok || got.TraceID != "trace" || got.RequestID != "request" || got.Route != "/route" {
		t.Fatalf("values=%#v ok=%v", got, ok)
	}
}

func TestFromMissingAndNilContext(t *testing.T) {
	for _, ctx := range []context.Context{nil, context.Background()} {
		if got, ok := From(ctx); ok || got != (Values{}) {
			t.Fatalf("values=%#v ok=%v", got, ok)
		}
	}
}
