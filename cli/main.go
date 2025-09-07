package main

import (
	"fmt"

	"github.com/adrg/xdg"
)

func loadConfig() {
	fastmaskConfigDir := xdg.ConfigHome + "/fastmask"
	fmt.Println("config dirs: ", fastmaskConfigDir)

	tokenFile := fastmaskConfigDir + "/config.json"
	fmt.Println("tokenFile: ", tokenFile)
}

func main() {
	loadConfig()
}
