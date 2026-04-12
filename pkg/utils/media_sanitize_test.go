package utils

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"plain":     {in: "image.png", want: "image.png"},
		"simple up": {in: "../foo.txt", want: "foo.txt"},
		"nested":    {in: ".//.", want: ""},
		"absolute":  {in: "/etc/passwd", want: "passwd"},
		"multi up":  {in: "../../a/b.txt", want: "b.txt"},
		"empty":     {in: "", want: ""},
		"dots":      {in: "..", want: ""},
		"mid dots":  {in: "test..txt", want: "test..txt"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := SanitizeFilename(tc.in); got != tc.want {
				t.Fatalf("SanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
