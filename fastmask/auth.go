package main

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
)

func Authenticate(httpClient *http.Client, token SecureToken) (fastmailIdentity *FastmailIdentity, err error) {
	req, err := http.NewRequest("GET", _authUrl, nil)
	if err != nil {
		return
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+token.FullToken())

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	if resp == nil {
		err = fmt.Errorf("strangely nil resp despite err being nil")
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			err = fmt.Errorf("failed to close response body: %w", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	body := string(bodyBytes)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", body)
		return
	}

	var fastmailAuthResponse FastmailAuthResponse
	err = json.Unmarshal([]byte(body), &fastmailAuthResponse)
	if err != nil {
		return
	}
	_, ok := fastmailAuthResponse.Capabilities[_maskedEmailCapability]
	if !ok {
		err = fmt.Errorf("fastmail token does not have masked email capability")
		return
	}

	accountId := fastmailAuthResponse.PrimaryAccounts[_primaryAccountKey]
	apiUrl := fastmailAuthResponse.APIURL
	accountIdStr, ok := accountId.(string)
	if !ok {
		err = fmt.Errorf("found account id but it was of unexpected type: %T", accountId)
		return
	}
	fastmailIdentity = &FastmailIdentity{
		APIURL:    apiUrl,
		AccountID: accountIdStr,
	}

	return
}
