package main

import (
	"github.com/rs/zerolog/log"
	"ip-syncer/src/control"
	"ip-syncer/src/icanhazip"
	"ip-syncer/src/transip"
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

	// Initialize the TransIP API client and repositories.
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
	}
	log.Debug().Msgf("IPv4: %s, IPv6: %s", ipAddresses.V4, ipAddresses.V6)

	// Loop domains
	for _, domainName := range a.config.Domains {
		log.Info().Msgf("→ Processing domainName: %s", domainName)

		records, dnsGetErr := a.transIpApi.GetDnsRecords(domainName)
		if dnsGetErr != nil {
			log.Error().Msgf("[Error]: %v\n", dnsGetErr)
		}

		updateErr := a.transIpApi.UpdateARecordsWithGivenIps(domainName, ipAddresses, records)
		if updateErr != nil {
			log.Error().Msgf("[Error]: %v\n", updateErr)
		}
	}
}
