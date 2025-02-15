package transip

import (
	"fmt"
	"strings"

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
	dnsEntries, repoErr := c.DomainRepo.GetDNSEntries(domainName)
	if repoErr != nil {
		return nil, fmt.Errorf("failed to retrieve DNS entries: %v", repoErr)
	}

	return dnsEntries, nil
}

func (c *Client) UpdateARecordsWithGivenIps(domainName, rootDomain string, ips *icanhazip.Ip, dnsEntries []domain.DNSEntry) error {
	var updatedEntries []domain.DNSEntry
	isRootDomain := domainName == rootDomain
	subdomain := "@"

	if !isRootDomain {
		subdomain = strings.TrimSuffix(domainName, "."+rootDomain)
	}

	for _, dnsEntry := range dnsEntries {
		// Only update if it's the correct entry (@ for root, subdomain otherwise)
		if dnsEntry.Name != subdomain {
			continue
		}

		if dnsEntry.Type == "A" && dnsEntry.Content != ips.V4 {
			log.Warn().Msgf(" → Updating A record [%s] from %s to %s", dnsEntry.Name, dnsEntry.Content, ips.V4)
			dnsEntry.Content = ips.V4
			updatedEntries = append(updatedEntries, dnsEntry)
		}

		if dnsEntry.Type == "AAAA" && dnsEntry.Content != ips.V6 {
			log.Warn().Msgf(" → Updating AAAA record [%s] from %s to %s", dnsEntry.Name, dnsEntry.Content, ips.V6)
			dnsEntry.Content = ips.V6
			updatedEntries = append(updatedEntries, dnsEntry)
		}
	}

	if len(updatedEntries) > 0 {
		updateErr := c.DomainRepo.ReplaceDNSEntries(rootDomain, updatedEntries)
		if updateErr != nil {
			return fmt.Errorf("failed to update DNS records: %v", updateErr)
		}
		log.Info().Msgf("✓ Successfully updated DNS records for %s", domainName)
	} else {
		log.Info().Msgf("✓ No changes needed for %s", domainName)
	}

	return nil
}
