package utils

import "testing"

func TestIsAudioFile(t *testing.T) {
	tests := map[string]struct {
		filename    string
		contentType string
		want        bool
	}{
		"extension match ignores case":               {filename: "VOICE.MP3", want: true},
		"content type prefix ignores case":           {contentType: "AuDiO/WAV; charset=utf-8", want: true},
		"ogg application exact match ignores case":   {contentType: "APPLICATION/OGG", want: true},
		"x-ogg application exact match ignores case": {contentType: "application/X-OGG", want: true},
		"non audio extension":                        {filename: "image.png", want: false},
		"non audio content type":                     {contentType: "text/plain", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsAudioFile(tc.filename, tc.contentType); got != tc.want {
				t.Fatalf("IsAudioFile(%q, %q) = %v, want %v", tc.filename, tc.contentType, got, tc.want)
			}
		})
	}
}
