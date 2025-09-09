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
	_authUrl = "https://api.fastmail.com/.well-known/jmap"

	_maskedEmailCapability = "https://www.fastmail.com/dev/maskedemail"
	_primaryAccountKey     = "urn:ietf:params:jmap:core"

	_httpTimeout = 5 * time.Second

	_maskedEmailSetMethod = "MaskedEmail/set"

	_jmapStateEnabled = "enabled"

	_fastmaskRequestId = "fastmask"
	_jmapCallId        = "0"

	_contentTypeHeader = "Content-Type"
	_authHeader        = "Authorization"
	_jsonContentType   = "application/json"
)

type SecureToken string

func (t SecureToken) String() string {
	if len(t) == 0 {
		return "<empty>"
	}
	if len(t) <= 8 {
		return "<redacted>"
	}
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
	MethodResponses []jsontext.Value `json:"methodResponses"`
}

type MaskedEmailSetResponseDetail struct {
	Email string `json:"email"`
}

type MaskedEmailSetResponse struct {
	Created map[string]MaskedEmailSetResponseDetail `json:"created"`
}

type FastmaskResponse struct {
	Prefix string `json:"prefix"`
	Domain string `json:"domain"`
	Email  string `json:"email"`
}

func loadToken() (SecureToken, error) {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	err := checkDirectoryPermissions(fastmaskConfigDir)
	if err != nil {
		return "", fmt.Errorf("directory permissions: %w", err)
	}

	tokenFile := fastmaskConfigDir + "/config.json"
	err = checkFilePermissions(tokenFile)
	if err != nil {
		return "", fmt.Errorf("file permissions: %w", err)
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

func auth(httpClient *http.Client, token SecureToken) (*FastmailIdentity, error) {
	req, err := http.NewRequest("GET", _authUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+string(token))

	resp, err := httpClient.Do(req)
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

func createMaskedEmail(httpClient *http.Client, fastmailId *FastmailIdentity, domain string, token SecureToken) (*FastmaskResponse, error) {
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
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	req, err := http.NewRequest("POST", fastmailId.APIURL, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+string(token))
	resp, err := httpClient.Do(req)
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
	if resp.StatusCode != 200 {
		body := string(bodyBytes)
		return nil, fmt.Errorf("server returned status code %d: %s", resp.StatusCode, body)
	}

	var fastmailResponse FastmailMaskedEmailResponse
	err = json.Unmarshal(bodyBytes, &fastmailResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(fastmailResponse.MethodResponses) < 1 {
		return nil, fmt.Errorf("empty method responses")
	}

	// Parse the heterogeneous array [method_name, response_object, call_id]
	var methodResponse []jsontext.Value
	err = json.Unmarshal(fastmailResponse.MethodResponses[0], &methodResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse method response array: %w", err)
	}

	if len(methodResponse) < 2 {
		return nil, fmt.Errorf("invalid method response length: %d", len(methodResponse))
	}

	// Now unmarshal just the response object (index 1)
	var setResponse MaskedEmailSetResponse
	err = json.Unmarshal(methodResponse[1], &setResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse set response: %w", err)
	}

	maskedEmail, exists := setResponse.Created[_fastmaskRequestId]
	if !exists {
		return nil, fmt.Errorf("no fastmask response in created map")
	}

	if maskedEmail.Email == "" {
		return nil, fmt.Errorf("email was empty in response")
	}

	return &FastmaskResponse{
		Prefix: prefix,
		Domain: domain,
		Email:  maskedEmail.Email,
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

	httpClient := &http.Client{
		Timeout: _httpTimeout,
	}

	fastmailId, err := auth(httpClient, token)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	fastmaskResponse, err := createMaskedEmail(httpClient, fastmailId, domain, token)
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
