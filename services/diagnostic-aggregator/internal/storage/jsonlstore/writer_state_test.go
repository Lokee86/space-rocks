package jsonlstore

import (
	"os"
	"testing"
)

func TestWriterCloseReturnsStableFirstResult(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "writer-")
	if err != nil {
		t.Fatal(err)
	}
	writer := newWriter(file, 0)
	first := writer.close()
	second := writer.close()
	if first != second {
		t.Fatalf("close results changed: first=%v second=%v", first, second)
	}
}
