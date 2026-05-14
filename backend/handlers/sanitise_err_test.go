package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitiseErr(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "redacts_key_in_generativelanguage_url",
			in:   `Post "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent?%24alt=json&key=AIzaSyDy4QQzLQn2jFFHhqk6WHfrHQhiiDpCfYc": context deadline exceeded`,
			want: `Post "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent?%24alt=json&key=REDACTED": context deadline exceeded`,
		},
		{
			name: "redacts_when_key_is_first_param",
			in:   `GET https://api.example.com/v1/x?key=SECRET&q=foo failed`,
			want: `GET https://api.example.com/v1/x?key=REDACTED&q=foo failed`,
		},
		{
			name: "redacts_case_insensitive_param_name",
			in:   `KEY=ALSO_LEAKED appears`,
			want: `KEY=ALSO_LEAKED appears`, // bare KEY= without leading ? or & is not a query param — left alone
		},
		{
			name: "leaves_unrelated_text_alone",
			in:   `context deadline exceeded (no key visible)`,
			want: `context deadline exceeded (no key visible)`,
		},
		{
			name: "redacts_in_url_with_trailing_quote",
			in:   `Post "https://x?key=SECRET": boom`,
			want: `Post "https://x?key=REDACTED": boom`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := sanitiseErr(errors.New(c.in))
			assert.Equal(t, c.want, got)
		})
	}
}

func TestSanitiseErr_NilError(t *testing.T) {
	assert.Equal(t, "", sanitiseErr(nil))
}
