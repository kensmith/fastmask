package main

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/adrg/xdg"
)

const (
	_authUrl = "https://api.fastmail.com/.well-known/jmap"
)

type Config struct {
	Token string `json:"token"`
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

func auth(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", _authUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("Body: ", string(body[:]))
	fmt.Println("Status code: ", resp.StatusCode)

	return nil
}

func main() {
	token, err := loadToken()
	if err != nil {
		panic(err)
	}
	err = auth(token)
	if err != nil {
		panic(err)
	}
}
