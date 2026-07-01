package quantize

import (
	"math"
	"testing"
)

func TestEncodeDecodeFloatGeneric(t *testing.T) {
	policy, ok := PolicyByName(PolicyFloatGeneric)
	if !ok {
		t.Fatal("missing float_generic policy")
	}

	encoded, err := EncodeFloat(policy, 12.3456)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != 12346 {
		t.Fatalf("encoded = %d, want 12346", encoded)
	}

	decoded, err := DecodeFloat(policy, encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !closeEnough(decoded, 12.346, 0.0001) {
		t.Fatalf("decoded = %v, want about 12.346", decoded)
	}
}

func TestRatioPolicyClampsAndDecodes(t *testing.T) {
	policy, ok := PolicyByName(PolicyRatio01)
	if !ok {
		t.Fatal("missing ratio_0_1 policy")
	}

	tests := []struct {
		name     string
		value    float64
		wantEnc  int64
		wantDec  float64
		tolerance float64
	}{
		{name: "zero", value: 0.0, wantEnc: 0, wantDec: 0.0, tolerance: 0.00001},
		{name: "half", value: 0.5, wantEnc: 32768, wantDec: 0.5000076295, tolerance: 0.00002},
		{name: "one", value: 1.0, wantEnc: 65535, wantDec: 1.0, tolerance: 0.00001},
		{name: "below", value: -0.25, wantEnc: 0, wantDec: 0.0, tolerance: 0.00001},
		{name: "above", value: 1.25, wantEnc: 65535, wantDec: 1.0, tolerance: 0.00001},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeFloat(policy, tc.value)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if encoded != tc.wantEnc {
				t.Fatalf("encoded = %d, want %d", encoded, tc.wantEnc)
			}

			decoded, err := DecodeFloat(policy, encoded)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !closeEnough(decoded, tc.wantDec, tc.tolerance) {
				t.Fatalf("decoded = %v, want about %v", decoded, tc.wantDec)
			}
		})
	}
}

func TestSecondsPolicyClampsAndDoesNotWrap(t *testing.T) {
	policy, ok := PolicyByName(PolicySeconds)
	if !ok {
		t.Fatal("missing seconds policy")
	}

	encoded, err := EncodeFloat(policy, 1.234)
	if err != nil {
		t.Fatalf("encode positive: %v", err)
	}
	if encoded != 1234 {
		t.Fatalf("encoded = %d, want 1234", encoded)
	}

	decoded, err := DecodeFloat(policy, encoded)
	if err != nil {
		t.Fatalf("decode positive: %v", err)
	}
	if !closeEnough(decoded, 1.234, 0.0001) {
		t.Fatalf("decoded = %v, want about 1.234", decoded)
	}

	encoded, err = EncodeFloat(policy, -2.5)
	if err != nil {
		t.Fatalf("encode negative clamp: %v", err)
	}
	if encoded != 0 {
		t.Fatalf("encoded = %d, want 0", encoded)
	}

	decoded, err = DecodeFloat(policy, encoded)
	if err != nil {
		t.Fatalf("decode negative clamp: %v", err)
	}
	if decoded != 0 {
		t.Fatalf("decoded = %v, want 0", decoded)
	}
}

func TestClampDoesNotWrapAtEncodedBounds(t *testing.T) {
	ratio, _ := PolicyByName(PolicyRatio01)
	seconds, _ := PolicyByName(PolicySeconds)

	tests := []struct {
		name   string
		policy Policy
		value  float64
		want   int64
	}{
		{name: "ratio below", policy: ratio, value: -100, want: 0},
		{name: "ratio above", policy: ratio, value: 100, want: 65535},
		{name: "seconds below", policy: seconds, value: -100, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeFloat(tc.policy, tc.value)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if encoded != tc.want {
				t.Fatalf("encoded = %d, want %d", encoded, tc.want)
			}
		})
	}
}

func TestEncodeFloatRejectsInvalidValues(t *testing.T) {
	policy, ok := PolicyByName(PolicyFloatGeneric)
	if !ok {
		t.Fatal("missing float_generic policy")
	}

	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := EncodeFloat(policy, value); err == nil {
			t.Fatalf("expected error for %v", value)
		}
	}
}

func closeEnough(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
