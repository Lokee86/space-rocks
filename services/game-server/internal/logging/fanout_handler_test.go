package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestFanoutHandlerWritesOneRecordToTwoHandlers(t *testing.T) {
	var first bytes.Buffer
	var second bytes.Buffer

	handler := fanoutHandler{
		handlers: []slog.Handler{
			slog.NewJSONHandler(&first, nil),
			slog.NewJSONHandler(&second, nil),
		},
	}

	logger := slog.New(handler)
	logger.InfoContext(context.Background(), "fanout works", "key", "value")

	firstOutput := first.String()
	secondOutput := second.String()

	if firstOutput == "" {
		t.Fatal("expected first handler to receive output")
	}

	if secondOutput == "" {
		t.Fatal("expected second handler to receive output")
	}

	for index, output := range []string{firstOutput, secondOutput} {
		if !strings.Contains(output, "\"msg\":\"fanout works\"") {
			t.Fatalf("expected handler %d output to include message, got %q", index+1, output)
		}

		if !strings.Contains(output, "\"key\":\"value\"") {
			t.Fatalf("expected handler %d output to include attribute, got %q", index+1, output)
		}
	}
}
