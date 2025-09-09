package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "time/tzdata"
)

func usage(programName string) {
	fmt.Printf("%s <domain>\n", programName)
}

func main() {
	if len(os.Args) != 2 {
		programName := "fastmask"
		if len(os.Args) > 0 {
			programName = os.Args[0]
		}
		usage(programName)
		os.Exit(1)
	}
	domain := os.Args[1]
	token, err := LoadToken()
	if err != nil {
		log.Fatalf(`
Failed to load token file: %v
Generate your token at: https://app.fastmail.com/settings/security/tokens
Then create the config file with:
{
  "token": "<your fastmail API token>"
}
Make sure to chmod 700 the directory and 600 the config file to protect your token
`, err)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: _httpTimeout,
	}

	fastmailId, err := Authenticate(httpClient, token)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	fastmaskResponse, err := CreateMaskedEmail(httpClient, fastmailId, domain, token)
	if err != nil {
		log.Fatalf("masked email: %v", err)
	}

	fastmaskResponseJSON, err := json.Marshal(fastmaskResponse, jsontext.WithIndent("  "))
	if err != nil {
		log.Fatalf("response marshal: %v", err)
	}

	fastmaskResponseStr := string(fastmaskResponseJSON)
	fmt.Println(fastmaskResponseStr)
}
