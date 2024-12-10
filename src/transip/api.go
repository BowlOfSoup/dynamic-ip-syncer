package transip

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/transip/gotransip/v6"
	"github.com/transip/gotransip/v6/domain"
	"ip-syncer/src/icanhazip"
)

type Account struct {
	AccountName    string
	PrivateKeyFile string
}

type Client struct {
	DomainRepo *domain.Repository
}

func InitClient(account Account) (*Client, error) {
	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		AccountName:    account.AccountName,
		PrivateKeyPath: account.PrivateKeyFile,
		TokenCache:     NewTokenCache(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create TransIP client: %v", err)
	}

	return &Client{
		DomainRepo: &domain.Repository{Client: client},
	}, nil
}

func (c *Client) GetDnsRecords(domainName string) ([]domain.DNSEntry, error) {
	// Retrieve the current DNS entries for the domain
	dnsEntries, repoErr := c.DomainRepo.GetDNSEntries(domainName)
	if repoErr != nil {
		return nil, fmt.Errorf("failed to retrieve DNS entries: %v", repoErr)
	}

	return dnsEntries, nil
}

func (c *Client) UpdateARecordsWithGivenIps(domainName string, ips *icanhazip.Ip, dnsEntries []domain.DNSEntry) error {
	for i, dnsEntry := range dnsEntries {
		if dnsEntry.Type == "A" {
			log.Debug().Msgf(" → Current A record: %s", dnsEntry.Content)
			if dnsEntry.Content != ips.V4 {
				log.Warn().Msgf(" → Updating A record from %s to %s", dnsEntry.Content, ips.V4)
				dnsEntries[i].Content = ips.V4
			}
		}
		if dnsEntry.Type == "AAAA" {
			log.Debug().Msgf(" → Current AAAA record: %s", dnsEntry.Content)
			if dnsEntry.Content != ips.V6 {
				log.Warn().Msgf(" → Updating AAAA record from %s to %s", dnsEntry.Content, ips.V6)
				dnsEntries[i].Content = ips.V6
			}
		}
	}

	updateErr := c.DomainRepo.ReplaceDNSEntries(domainName, dnsEntries)
	if updateErr != nil {
		return updateErr
	}

	return nil
}
