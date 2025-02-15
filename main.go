package main

import (
	"github.com/rs/zerolog/log"
	"ip-syncer/src/control"
	"ip-syncer/src/icanhazip"
	"ip-syncer/src/transip"
	"strings"
	"time"
)

type appState struct {
	config     *control.Config
	transIpApi *transip.Client
}

func main() {
	control.CaptureSigTerm()
	control.InitLogger()
	log.Info().Msg("++ Starting IP syncer ++")

	// Get configuration.
	var config control.Config
	if err := control.LoadConfig(&config); err != nil {
		log.Fatal().Msgf("Failed to load configuration: %v", err)
	}

	transIpApi, clientErr := transip.InitClient(
		transip.Account{
			AccountName:    config.Account.Name,
			PrivateKeyFile: config.Account.PrivateKeyPath,
		},
	)
	if clientErr != nil {
		log.Fatal().Msgf("[Error] [TransIP]: %v\n", clientErr)
	}

	app := &appState{
		config:     &config,
		transIpApi: transIpApi,
	}

	// --- Main loop --
	ticker := time.NewTicker(time.Duration(config.SyncInterval) * time.Second)
	defer ticker.Stop()

	app.syncIpAddresses()
	for {
		select {
		case <-ticker.C:
			app.syncIpAddresses()
		}
	}
}

func (a *appState) syncIpAddresses() {
	// Fetch current IP Addresses.
	ipAddresses, ipErr := icanhazip.GetIPAddresses()
	if ipErr != nil {
		log.Error().Msgf("[Error] [icanhazip]: %v\n", ipErr)
		return
	}
	log.Debug().Msgf("IPv4: %s, IPv6: %s", ipAddresses.V4, ipAddresses.V6)

	// Determine root domains
	rootDomains := make(map[string]bool)
	for _, domainName := range a.config.Domains {
		parts := strings.Split(domainName, ".")
		if len(parts) == 2 { // This is a root domain
			rootDomains[domainName] = true
		}
	}

	for _, domainName := range a.config.Domains {
		log.Info().Msgf("→ Processing domainName: %s", domainName)

		// Identify root domain for this domain
		rootDomain := findRootDomain(domainName, rootDomains)
		if rootDomain == "" {
			log.Error().Msgf("[Error] Could not determine root domain for %s", domainName)
			continue
		}

		records, dnsGetErr := a.transIpApi.GetDnsRecords(rootDomain)
		if dnsGetErr != nil {
			log.Error().Msgf("[Error]: %v\n", dnsGetErr)
			continue
		}

		// Update records correctly (root `@` or specific subdomain)
		updateErr := a.transIpApi.UpdateARecordsWithGivenIps(domainName, rootDomain, ipAddresses, records)
		if updateErr != nil {
			log.Error().Msgf("[Error]: %v\n", updateErr)
		}
	}
}

// findRootDomain finds the root domain from a given domain name
func findRootDomain(domainName string, rootDomains map[string]bool) string {
	parts := strings.Split(domainName, ".")
	if len(parts) < 2 {
		return ""
	}

	for i := 0; i < len(parts)-1; i++ {
		rootCandidate := strings.Join(parts[i:], ".")
		if rootDomains[rootCandidate] {
			return rootCandidate
		}
	}

	return ""
}
