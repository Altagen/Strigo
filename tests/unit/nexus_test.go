package unit

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Path with spaces",
			input:    "jdk/my distribution/temurin",
			expected: "jdk%2Fmy+distribution%2Ftemurin",
		},
		{
			name:     "Path with ampersand (injection attempt)",
			input:    "jdk&token=stolen",
			expected: "jdk%26token%3Dstolen",
		},
		{
			name:     "Path with question mark",
			input:    "jdk?admin=true",
			expected: "jdk%3Fadmin%3Dtrue",
		},
		{
			name:     "Normal path",
			input:    "jdk/adoptium/temurin",
			expected: "jdk%2Fadoptium%2Ftemurin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := url.QueryEscape(tt.input)
			assert.Equal(t, tt.expected, encoded)
		})
	}
}
