package utils

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "normal filename",
			filename: "image.png",
			want:     "image.png",
		},
		{
			name:     "path traversal - simple up",
			filename: "../foo.txt",
			want:     "foo.txt",
		},
		{
			name:     "path traversal - nested",
			filename: ".//.",
			want:     "",
		},
		{
			name:     "path traversal - absolute",
			filename: "/etc/passwd",
			want:     "passwd",
		},
		{
			name:     "path traversal - multiple up",
			filename: "../../../../../etc/passwd",
			want:     "passwd",
		},
		{
			name:     "empty",
			filename: "",
			want:     "",
		},
		{
			name:     "dots",
			filename: "..",
			want:     "",
		},
		{
			name:     "double dots in middle of name",
			filename: "test..txt",
			want:     "test..txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeFilename(tt.filename); got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
