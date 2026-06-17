package money

import (
	"errors"
	"testing"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Price
		wantErr bool
	}{
		{
			name:  "whole amount",
			input: "12",
			want:  Price(12_000_000),
		},
		{
			name:  "single fractional digit",
			input: "12.3",
			want:  Price(12_300_000),
		},
		{
			name:  "maximum supported fractional precision",
			input: "12.345678",
			want:  Price(12_345_678),
		},
		{
			name:  "smallest representable fraction",
			input: "0.000001",
			want:  Price(1),
		},
		{
			name:  "surrounding whitespace",
			input: " 12.34 ",
			want:  Price(12_340_000),
		},
		{
			name:    "empty value",
			input:   "",
			wantErr: true,
		},
		{
			name:    "negative value",
			input:   "-1.00",
			wantErr: true,
		},
		{
			name:    "too many fractional digits",
			input:   "12.3456789",
			wantErr: true,
		},
		{
			name:    "comma decimal separator",
			input:   "12,34",
			wantErr: true,
		},
		{
			name:    "exponent notation",
			input:   "1e3",
			wantErr: true,
		},
		{
			name:    "missing whole number",
			input:   ".50",
			wantErr: true,
		},
		{
			name:    "missing fractional number",
			input:   "12.",
			wantErr: true,
		},
		{
			name:    "invalid suffix",
			input:   "12.34abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrice(tt.input)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPrice) {
					t.Fatalf("ParsePrice(%q) error = %v, want ErrInvalidPrice", tt.input, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePrice(%q) unexpected error = %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("ParsePrice(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
