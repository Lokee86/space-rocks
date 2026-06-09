package codec

import "testing"

func TestDecodeType(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		got, err := DecodeType([]byte(`{"type":"player_data_load_stats"}`))
		if err != nil {
			t.Fatalf("DecodeType returned error: %v", err)
		}
		if got != "player_data_load_stats" {
			t.Fatalf("DecodeType returned %q, want %q", got, "player_data_load_stats")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		if _, err := DecodeType([]byte(`{"type":`)); err == nil {
			t.Fatal("DecodeType returned nil error for malformed JSON")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		if _, err := DecodeType([]byte(`{"message":"ok"}`)); err == nil {
			t.Fatal("DecodeType returned nil error for missing type")
		}
	})

	t.Run("empty type", func(t *testing.T) {
		if _, err := DecodeType([]byte(`{"type":""}`)); err == nil {
			t.Fatal("DecodeType returned nil error for empty type")
		}
	})
}
