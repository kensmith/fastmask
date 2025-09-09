package main

import (
	"encoding/json/v2"
	"fmt"
	"os"

	"github.com/adrg/xdg"
)

type SecureToken struct {
	token string
}

func NewSecureToken(t string) SecureToken {
	return SecureToken{
		token: t,
	}
}

func (st SecureToken) String() string {
	if len(st.token) == 0 {
		return "<empty>"
	}
	if len(st.token) <= 8 {
		return "<redacted>"
	}
	return string(st.token[:4]) + "..." + "<redacted>"
}

func (st SecureToken) FullToken() string {
	return st.token
}

func (st SecureToken) Equals(s string) bool {
	return st.token == s
}

func LoadToken() (SecureToken, error) {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	err := checkDirectoryPermissions(fastmaskConfigDir)
	if err != nil {
		return SecureToken{}, fmt.Errorf("directory permissions: %w", err)
	}

	tokenFile := fastmaskConfigDir + "/config.json"
	err = checkFilePermissions(tokenFile)
	if err != nil {
		return SecureToken{}, fmt.Errorf("file permissions: %w", err)
	}

	tokenData, err := os.ReadFile(tokenFile)
	if err != nil {
		return SecureToken{}, err
	}
	var config Config
	err = json.Unmarshal(tokenData, &config)
	if err != nil {
		return SecureToken{}, err
	}
	if config.Token == "" {
		return SecureToken{}, fmt.Errorf("token is empty in config file")
	}
	return NewSecureToken(config.Token), nil
}
