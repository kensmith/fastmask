package main

import (
	"context"
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := string(bodyBytes[:])

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", body)
	}

	fmt.Println(body)

	return nil
}

func main() {
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
	err = auth(token)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
}
