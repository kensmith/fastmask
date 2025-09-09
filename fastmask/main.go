package main

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/adrg/xdg"
)

const (
	// API endpoints
	_authUrl = "https://api.fastmail.com/.well-known/jmap"

	// JMAP capabilities
	_maskedEmailCapability = "https://www.fastmail.com/dev/maskedemail"
	_primaryAccountKey     = "urn:ietf:params:jmap:core"

	// HTTP configuration
	_httpTimeout = 5 * time.Second

	// File permissions
	_configDirPerm          = 0o700
	_configFilePerm         = 0o600
	_configFilePermReadOnly = 0o400

	// JMAP method names
	_maskedEmailSetMethod = "MaskedEmail/set"

	// JMAP response keys
	_jmapCreatedKey   = "created"
	_jmapEmailKey     = "email"
	_jmapStateEnabled = "enabled"

	// Request identifiers
	_fastmaskRequestId = "fastmask"
	_jmapCallId        = "0"

	// Headers
	_contentTypeHeader = "Content-Type"
	_authHeader        = "Authorization"
	_jsonContentType   = "application/json"
)

var defaultClient = &http.Client{
	Timeout: _httpTimeout,
}

type SecureToken string

func (t SecureToken) String() string {
	if len(t) == 0 {
		return "<empty>"
	}
	if len(t) <= 8 {
		return "<redacted>"
	}
	// Show only first 4 chars for debugging, rest is masked
	return string(t[:4]) + "..." + "<redacted>"
}

type Config struct {
	Token SecureToken `json:"token"`
}

type FastmailIdentity struct {
	AccountID string
	APIURL    string
}

type FastmailAuthResponse struct {
	APIURL          string         `json:"apiUrl"`
	Capabilities    map[string]any `json:"capabilities"`
	PrimaryAccounts map[string]any `json:"primaryAccounts"`
}

type FastmailMaskedEmailCreateMethodParameters struct {
	ForDomain   string `json:"forDomain"`
	State       string `json:"state"`
	EmailPrefix string `json:"emailPrefix"`
}

type FastmailMaskedEmailCreateMethod struct {
	AccountId string                                               `json:"accountId"`
	Create    map[string]FastmailMaskedEmailCreateMethodParameters `json:"create"`
}

type FastmailMaskedEmailRequest struct {
	Using       []string `json:"using"`
	MethodCalls [][]any  `json:"methodCalls"`
}

type FastmailMaskedEmailResponse struct {
	MethodResponses [][]any `json:"methodResponses"`
}

type FastmaskResponse struct {
	Prefix string `json:"prefix"`
	Domain string `json:"domain"`
	Email  string `json:"email"`
}

func loadToken() (SecureToken, error) {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	tokenFile := fastmaskConfigDir + "/config.json"

	// Check directory permissions
	dirInfo, err := os.Stat(fastmaskConfigDir)
	if err != nil {
		return "", fmt.Errorf("cannot access config directory: %w", err)
	}
	dirMode := dirInfo.Mode().Perm()
	if dirMode != _configDirPerm {
		return "", fmt.Errorf("insecure config directory permissions %04o (should be %04o): %s", dirMode, _configDirPerm, fastmaskConfigDir)
	}

	// Check file permissions
	fileInfo, err := os.Stat(tokenFile)
	if err != nil {
		return "", fmt.Errorf("cannot access config file: %w", err)
	}
	fileMode := fileInfo.Mode().Perm()
	if fileMode != _configFilePerm && fileMode != _configFilePermReadOnly {
		return "", fmt.Errorf("insecure config file permissions %04o (should be %04o or %04o): %s", fileMode, _configFilePerm, _configFilePermReadOnly, tokenFile)
	}

	tokenData, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	var config Config
	err = json.Unmarshal(tokenData, &config)
	if err != nil {
		return "", err
	}
	if config.Token == "" {
		return "", fmt.Errorf("token is empty in config file")
	}
	return config.Token, nil
}

func auth(token SecureToken) (*FastmailIdentity, error) {
	req, err := http.NewRequest("GET", _authUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+string(token))

	resp, err := defaultClient.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatalf("failed to close response body: %v", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", body)
	}

	var fastmailAuthResponse FastmailAuthResponse
	err = json.Unmarshal([]byte(body), &fastmailAuthResponse)
	if err != nil {
		return nil, err
	}
	_, ok := fastmailAuthResponse.Capabilities[_maskedEmailCapability]
	if !ok {
		return nil, fmt.Errorf("fastmail token does not have masked email capability")
	}

	accountId := fastmailAuthResponse.PrimaryAccounts[_primaryAccountKey]
	apiUrl := fastmailAuthResponse.APIURL
	accountIdStr, ok := accountId.(string)
	if !ok {
		return nil, fmt.Errorf("found account id but it was of unexpected type: %T", accountId)
	}

	return &FastmailIdentity{
		APIURL:    apiUrl,
		AccountID: accountIdStr,
	}, nil
}

func createMaskedEmail(fastmailId *FastmailIdentity, domain string, token SecureToken) (*FastmaskResponse, error) {
	prefix := GenPrefix()
	var request FastmailMaskedEmailRequest
	request.Using = []string{_primaryAccountKey, _maskedEmailCapability}
	request.MethodCalls = [][]any{
		{
			_maskedEmailSetMethod,
			FastmailMaskedEmailCreateMethod{
				AccountId: fastmailId.AccountID,
				Create: map[string]FastmailMaskedEmailCreateMethodParameters{
					_fastmaskRequestId: {
						ForDomain:   domain,
						State:       _jmapStateEnabled,
						EmailPrefix: prefix,
					},
				},
			},
			_jmapCallId,
		},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %v", err)
	}

	req, err := http.NewRequest("POST", fastmailId.APIURL, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+string(token))
	resp, err := defaultClient.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatalf("failed to close response body: %v", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status code %d: %s", resp.StatusCode, body)
	}

	var fastmailResponse FastmailMaskedEmailResponse
	err = json.Unmarshal(bodyBytes, &fastmailResponse)
	if err != nil {
		return nil, err
	}

	// the following manual unmarshaling code is due to the jmap response array
	// containing both strings and maps which does not elegantly map to nested
	// Go structs
	responsesLength := len(fastmailResponse.MethodResponses)
	if responsesLength < 1 {
		return nil, fmt.Errorf("invalid responses length: %v", responsesLength)
	}
	response := fastmailResponse.MethodResponses[0]
	responseLength := len(response)
	if responseLength < 2 {
		return nil, fmt.Errorf("invalid response length: %v", responseLength)
	}

	responseMap, ok := response[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("response was not a map: %v", response[1])
	}

	createdAny, ok := responseMap[_jmapCreatedKey]
	if !ok {
		return nil, fmt.Errorf("%s key not found in response map: %v", _jmapCreatedKey, responseMap)
	}

	created, ok := createdAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("value for %s was not a map: %v", _jmapCreatedKey, createdAny)
	}

	fastmaskAny, ok := created[_fastmaskRequestId]
	if !ok {
		return nil, fmt.Errorf("no %s response in payload: %v", _fastmaskRequestId, created)
	}

	fastmaskMap, ok := fastmaskAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s response was not a map: %v", _fastmaskRequestId, fastmaskAny)
	}

	emailAny, ok := fastmaskMap[_jmapEmailKey]
	if !ok {
		return nil, fmt.Errorf("no %s in payload: %v", _jmapEmailKey, fastmaskMap)
	}

	email, ok := emailAny.(string)
	if !ok {
		return nil, fmt.Errorf("%s was not a string: %v", _jmapEmailKey, emailAny)
	}

	return &FastmaskResponse{
		Prefix: prefix,
		Domain: domain,
		Email:  email,
	}, nil
}

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
	token, err := loadToken()
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
	fastmailId, err := auth(token)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	fastmaskResponse, err := createMaskedEmail(fastmailId, domain, token)
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
