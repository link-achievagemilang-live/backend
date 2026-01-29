package service

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Zero", 0, "0"},
		{"One", 1, "1"},
		{"Ten", 10, "a"},
		{"Sixty-Two", 62, "10"},
		{"Large Number", 123456789, "8m0Kx"},
		{"Max Int32", 2147483647, "2lkCB1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Encode(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		hasError bool
	}{
		{"Zero", "0", 0, false},
		{"One", "1", 1, false},
		{"Ten", "a", 10, false},
		{"Sixty-Two", "10", 62, false},
		{"Large Number", "8m0Kx", 123456789, false},
		{"Max Int32", "2lkCB1", 2147483647, false},
		{"Invalid Character", "invalid!", 0, true},
		{"Empty String", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decode(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Decode(%s) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Decode(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Decode(%s) = %d; want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []int64{0, 1, 10, 62, 100, 1000, 10000, 100000, 1000000, 123456789}

	for _, num := range tests {
		t.Run("", func(t *testing.T) {
			encoded := Encode(num)
			decoded, err := Decode(encoded)
			if err != nil {
				t.Errorf("Decode failed for encoded value of %d: %v", num, err)
			}
			if decoded != num {
				t.Errorf("Round trip failed: %d -> %s -> %d", num, encoded, decoded)
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(123456789)
	}
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Decode("8m0Kx")
	}
}
