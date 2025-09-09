package main

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"log"
	"net/http"
)

func Authenticate(httpClient *http.Client, token SecureToken) (*FastmailIdentity, error) {
	req, err := http.NewRequest("GET", _authUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+token.FullToken())

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
