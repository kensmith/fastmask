package main

import "encoding/json/jsontext"

type Config struct {
	Token string `json:"token"`
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
