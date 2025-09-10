package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	_ "time/tzdata"
)

func usage(programName string) {
	fmt.Printf(`
usage:
	%s <domain>

Generate your token at: https://app.fastmail.com/settings/security/tokens
Then create the config file with:
{
  "token": "<your fastmail API token>"
}
Make sure to chmod 700 the directory and 600 the config file to protect your token
`, programName)
}

func main() {
	ConfigureLogging()
	programName := _programName
	if len(os.Args) != 2 {
		if len(os.Args) > 0 {
			programName = os.Args[0]
		}
		usage(programName)
		os.Exit(1)
	}
	domain := os.Args[1]
	token, err := LoadToken()
	if err != nil {
		slog.Error("token load", "err", err)
		usage(programName)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: _httpTimeout,
	}

	fastmailId, err := Authenticate(httpClient, token)
	if err != nil {
		slog.Error("authentication", "err", err)
		os.Exit(1)
	}

	fastmaskResponse, err := CreateMaskedEmail(httpClient, fastmailId, domain, token)
	if err != nil {
		slog.Error("masked email", "err", err)
		os.Exit(1)
	}

	fastmaskResponseJSON, err := json.Marshal(fastmaskResponse, jsontext.WithIndent("  "))
	if err != nil {
		slog.Error("response marshal", "err", err)
		os.Exit(1)
	}

	fastmaskResponseStr := string(fastmaskResponseJSON)
	fmt.Println(fastmaskResponseStr)
}
