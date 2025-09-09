package main

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"log"
	"net/http"
)

type FastmailIdentity struct {
	AccountID string
	APIURL    string
}

func CreateMaskedEmail(httpClient *http.Client, fastmailId *FastmailIdentity, domain string, token SecureToken) (*FastmaskResponse, error) {
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
