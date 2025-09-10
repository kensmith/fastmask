package main

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
)

type FastmailIdentity struct {
	AccountID string
	APIURL    string
}

func CreateMaskedEmail(httpClient *http.Client, fastmailId *FastmailIdentity, domain string, token SecureToken) (fastmaskResponse *FastmaskResponse, errReturned error) {
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
		err = fmt.Errorf("json marshal: %w", err)
		errReturned = err
		return
	}

	req, err := http.NewRequest("POST", fastmailId.APIURL, bytes.NewReader(jsonRequest))
	if err != nil {
		errReturned = err
		return
	}
	req.Header.Set(_contentTypeHeader, _jsonContentType)
	req.Header.Set(_authHeader, "Bearer "+token.FullToken())
	resp, err := httpClient.Do(req)
	if err != nil {
		errReturned = err
		return
	}
	if resp == nil {
		err = fmt.Errorf("strangely nil resp despite err being nil")
		errReturned = err
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			err = fmt.Errorf("close response body: %w", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read response body %d: %w", resp.StatusCode, err)
		errReturned = err
		return
	}

	if resp.StatusCode != 200 {
		body := string(bodyBytes)
		err = fmt.Errorf("http code %d: %s", resp.StatusCode, body)
		errReturned = err
		return
	}

	var fastmailResponse FastmailMaskedEmailResponse
	err = json.Unmarshal(bodyBytes, &fastmailResponse)
	if err != nil {
		err = fmt.Errorf("parse response: %w", err)
		errReturned = err
		return
	}

	if len(fastmailResponse.MethodResponses) < 1 {
		err = fmt.Errorf("empty method responses")
		errReturned = err
		return
	}

	// Parse the heterogeneous array [method_name, response_object, call_id]
	var methodResponse []jsontext.Value
	err = json.Unmarshal(fastmailResponse.MethodResponses[0], &methodResponse)
	if err != nil {
		err = fmt.Errorf("parse method response array: %w", err)
		errReturned = err
		return
	}

	if len(methodResponse) < 2 {
		err = fmt.Errorf("invalid method response length: %d", len(methodResponse))
		errReturned = err
		return
	}

	// Now unmarshal just the response object (index 1)
	var setResponse MaskedEmailSetResponse
	err = json.Unmarshal(methodResponse[1], &setResponse)
	if err != nil {
		err = fmt.Errorf("parse set response: %w", err)
		errReturned = err
		return
	}

	maskedEmail, exists := setResponse.Created[_fastmaskRequestId]
	if !exists {
		err = fmt.Errorf("no fastmask response in created map")
		errReturned = err
		return
	}

	if maskedEmail.Email == "" {
		err = fmt.Errorf("email was empty in response")
		errReturned = err
		return
	}

	fastmaskResponse = &FastmaskResponse{
		Prefix: prefix,
		Domain: domain,
		Email:  maskedEmail.Email,
	}

	errReturned = nil
	return
}
