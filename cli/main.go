package main

import (
	"encoding/json/v2"
	"fmt"
	"os"

	"github.com/adrg/xdg"
)

type Config struct {
	Token string `json:"token"`
}

func loadConfig() {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	tokenFile := fastmaskConfigDir + "/config.json"
	tokenData, err := os.ReadFile(tokenFile)
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(tokenData, &config)
	if err != nil {
		panic(err)
	}
	fmt.Println("token: ", config.Token)
}

func main() {
	loadConfig()
}
