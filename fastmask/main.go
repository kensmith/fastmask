package main

import (
	"bytes"
	"context"
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
	_authUrl           = "https://api.fastmail.com/.well-known/jmap"
	_maskedEmailUrl    = "https://www.fastmail.com/dev/maskedemail"
	_primaryAccountKey = "urn:ietf:params:jmap:core"
)

type Config struct {
	Token string `json:"token"`
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

type FastmailMaskedEmail struct {
	Email string `json:"email"`
}

type FastmailFastmaskCreated struct {
	Fastmask FastmailMaskedEmail `json:"fastmask"`
}

type FastmailMethodResponse struct {
	// Created map[string]FastmailFastmaskCreated `json:"created"`
	Created map[string]any `json:"created"`
}

type FastmailMaskedEmailResponse struct {
	MethodResponses [][]any `json:"methodResponses"`
}

type FastmaskResponse struct {
	Prefix string `json:"prefix"`
	Domain string `json:"domain"`
	Email  string `json:"email"`
}

func loadToken() (string, error) {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	tokenFile := fastmaskConfigDir + "/config.json"
	tokenData, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	var config Config
	err = json.Unmarshal(tokenData, &config)
	if err != nil {
		return "", err
	}
	return config.Token, nil
}

func auth(token string) (*FastmailIdentity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", _authUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes[:])

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", body)
	}

	var fastmailAuthResponse FastmailAuthResponse
	err = json.Unmarshal([]byte(body), &fastmailAuthResponse)
	if err != nil {
		return nil, err
	}
	_, ok := fastmailAuthResponse.Capabilities[_maskedEmailUrl]
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

func createMaskedEmail(fastmailId *FastmailIdentity, domain, token string) (*FastmaskResponse, error) {
	prefix := GenPrefix()
	var request FastmailMaskedEmailRequest
	request.Using = []string{_primaryAccountKey, _maskedEmailUrl}
	request.MethodCalls = [][]any{
		{
			"MaskedEmail/set",
			FastmailMaskedEmailCreateMethod{
				AccountId: fastmailId.AccountID,
				Create: map[string]FastmailMaskedEmailCreateMethodParameters{
					"fastmask": {
						ForDomain:   domain,
						State:       "enabled",
						EmailPrefix: prefix,
					},
				},
			},
			"0",
		},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", fastmailId.APIURL, bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes[:])

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

	createdAny, ok := responseMap["created"]
	if !ok {
		return nil, fmt.Errorf("created key not found in response map: %v", responseMap)
	}

	created, ok := createdAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("value for created was not a map: %v", createdAny)
	}

	fastmaskAny, ok := created["fastmask"]
	if !ok {
		return nil, fmt.Errorf("no fastmask response in payload: %v", created)
	}

	fastmaskMap, ok := fastmaskAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fastmask response was not a map: %v", fastmaskAny)
	}

	emailAny, ok := fastmaskMap["email"]
	if !ok {
		return nil, fmt.Errorf("no email in payload: %v", fastmaskMap)
	}

	email, ok := emailAny.(string)
	if !ok {
		return nil, fmt.Errorf("email was not a string: %v", emailAny)
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
	fastmaskResponseJSON, err := json.Marshal(fastmaskResponse)
	if err != nil {
		log.Fatalf("response marshal: %v", err)
	}
	_ = (*jsontext.Value)(&fastmaskResponseJSON).Indent()
	fastmaskResponseStr := string(fastmaskResponseJSON[:])
	fmt.Println(fastmaskResponseStr)
}
