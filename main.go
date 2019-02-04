package main

import "flag"

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	godaddy "github.com/kryptoslogic/godaddy-domainclient"
)

//import "github.com/davecgh/go-spew/spew"

var userDomainName = flag.String("domain", "", "Domain name")
var userIPAddress = flag.String("ip-address", "", "Set the top-level A record for this domain")
var applyBool = flag.Bool("apply", false, "Apply changes to domain")

func main() {
	flag.Parse()
	var apiConfig = godaddy.NewConfiguration()

	key := os.Getenv("GODADDY_KEY")
	secret := os.Getenv("GODADDY_SECRET")

	if net.ParseIP(*userIPAddress) == nil {
		fmt.Println("* The IP address provided to --ip-address is invalid")
		os.Exit(1)
	}

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

	// Check that the domain specified by the user exists in GoDaddy
	domainExists := 0
	for _, zone := range zones {
		if strings.EqualFold(zone.Status, "active") && strings.EqualFold(zone.Domain, *userDomainName) {
			domainExists = 1
			log.Println("Domain", zone.Domain, "matches user domain", *userDomainName)
		}
	}

	if domainExists != 1 {
		log.Println("Domain", *userDomainName, "does not match any domains returned from the API")
		os.Exit(1)
	}

	var recordData map[string]string
	recordData = make(map[string]string)

	records, _, err := apiClient.V1Api.RecordGet(ctx, *userDomainName, "", "", nil)
	for _, record := range records {
		if record.Name == "@" {
			if record.Type_ == "A" {
				log.Println("Changing", record.Type_, record.Data, "to", *userIPAddress)
				// Fill in the recordData map so it can be used outside this loop
				recordData["type"] = "A"
				recordData["name"] = "@"

			}
		}
		//log.Println(record.Name, "=>", record.Data)
	}

	if *applyBool {
		changeDataArray := []godaddy.DnsRecordCreateTypeName{{
			Data: *userIPAddress,
			Ttl:  600,
		}}

		result, _ := apiClient.V1Api.RecordReplaceTypeName(ctx, *userDomainName, "A", "@", changeDataArray, nil)

		if result.StatusCode == 200 {
			log.Println("Changed", recordData["type"], recordData["name"], "to", *userIPAddress)
			os.Exit(0)
		}
		if result.StatusCode != 200 {
			log.Println("Failed changing", recordData["type"], recordData["name"], "to", *userIPAddress)
			os.Exit(1)
		}

	} else {
		log.Println("The --apply bool is required to apply the changes")
		os.Exit(1)
	}
}
