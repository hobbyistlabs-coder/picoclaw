package api

import "testing"

func TestSanitizeSessionKey(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"normal":    {in: "agent:main:pico:direct:pico:1234", want: "agent_main_pico_direct_pico_1234"},
		"simple up": {in: "../foo", want: "foo"},
		"nested":    {in: ".//.", want: ""},
		"absolute":  {in: "/etc/passwd", want: "passwd"},
		"multi up":  {in: "../../a/b", want: "b"},
		"empty":     {in: "", want: ""},
		"dots":      {in: "..", want: ""},
		"colon mix": {in: "../foo:bar", want: "foo_bar"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := sanitizeSessionKey(tc.in); got != tc.want {
				t.Fatalf("sanitizeSessionKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
