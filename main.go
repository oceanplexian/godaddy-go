package main

import (
	"context"
	"fmt"
	"log"
	"os"

	godaddy "github.com/kryptoslogic/godaddy-domainclient"
)

func main() {
	fmt.Println("hello world")
	var apiConfig = godaddy.NewConfiguration()

	key := os.Getenv("GODADDY_KEY")
	secret := os.Getenv("GODADDY_SECRET")

	apiConfig.BasePath = "https://api.godaddy.com/"
	// Set auth
	//var authString = fmt.Sprintf("sso-key %s:%s", key, secret)
	var authString = "sso-key " + key + ":" + secret
	apiConfig.AddDefaultHeader("Authorization", authString)
	ctx := context.Background()

	var apiClient = godaddy.NewAPIClient(apiConfig)
	zones, resp, err := apiClient.V1Api.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	for _, zone := range zones {
		if zone.Status == "ACTIVE" {
			log.Println(zone.Domain)
		}
	}

	records, _, err := apiClient.V1Api.RecordGet(ctx, "lwts.org", "", "", nil)
	for _, record := range records {
		log.Println(record.Name, "=>", record.Data)
	}

	domainInfo, _, err := apiClient.V1Api.Get(ctx, "lwts.org", nil)
	log.Println(domainInfo)

	fmt.Print()
}
