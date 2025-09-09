package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestSecureTokenStringer(t *testing.T) {
	tests := []struct {
		name             string
		token            SecureToken
		expected         string
		shouldNotContain string
	}{
		{
			name:     "empty token",
			token:    SecureToken{},
			expected: "<empty>",
		},
		{
			name:     "short token",
			token:    NewSecureToken("abc"),
			expected: "<redacted>",
		},
		{
			name:             "normal token",
			token:            NewSecureToken("supersecrettoken123"),
			expected:         "supe...<redacted>",
			shouldNotContain: "secrettoken",
		},
		{
			name:             "API token example",
			token:            NewSecureToken("fmu1-abcd1234efgh5678ijkl9012mnop"),
			expected:         "fmu1...<redacted>",
			shouldNotContain: "abcd1234efgh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.String()
			if result != tt.expected {
				t.Errorf("SecureToken.String() = %v, want %v", result, tt.expected)
			}

			formatted := fmt.Sprintf("%v", tt.token)
			if formatted != tt.expected {
				t.Errorf("fmt.Sprintf(\"%%v\", token) = %v, want %v", formatted, tt.expected)
			}

			if tt.shouldNotContain != "" && strings.Contains(result, tt.shouldNotContain) {
				t.Errorf("SecureToken.String() exposed sensitive data: found %q in %q", tt.shouldNotContain, result)
			}
		})
	}
}

func TestSecureTokenExplicitConversion(t *testing.T) {
	originalToken := "fmu1-supersecretapitoken123456789"
	secureToken := NewSecureToken(originalToken)

	redacted := secureToken.String()
	if strings.Contains(redacted, "supersecret") {
		t.Errorf("String() method exposed sensitive data: %v", redacted)
	}
	if redacted != "fmu1...<redacted>" {
		t.Errorf("String() method returned unexpected value: %v", redacted)
	}

	formatted := fmt.Sprintf("%v", secureToken)
	if strings.Contains(formatted, "supersecret") {
		t.Errorf("fmt.Sprintf exposed sensitive data: %v", formatted)
	}

	fullToken := secureToken.FullToken()
	if fullToken != originalToken {
		t.Errorf("string() conversion failed to return full token\ngot:  %v\nwant: %v", fullToken, originalToken)
	}

	headerValue := "Bearer " + secureToken.FullToken()
	expectedHeader := "Bearer " + originalToken
	if headerValue != expectedHeader {
		t.Errorf("Header value would be incorrect\ngot:  %v\nwant: %v", headerValue, expectedHeader)
	}

	wrongHeaderViaFmt := fmt.Sprintf("Bearer %v", secureToken)
	if wrongHeaderViaFmt == expectedHeader {
		t.Errorf("fmt.Sprintf should NOT produce the correct header (it should be redacted)")
	}
	if !strings.Contains(wrongHeaderViaFmt, "<redacted>") {
		t.Errorf("fmt.Sprintf should produce redacted output, got: %v", wrongHeaderViaFmt)
	}
}
