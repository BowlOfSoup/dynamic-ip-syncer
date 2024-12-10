package icanhazip

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const ipv4Url = "https://ipv4.icanhazip.com/"
const ipv6Url = "https://ipv6.icanhazip.com/"

type Ip struct {
	V4 string
	V6 string
}

func fetchIpByUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IP from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			log.Fatalf("[canihazip] error: failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from %s: %w", url, err)
	}

	return strings.TrimSpace(string(body)), nil
}

func GetIPAddresses() (*Ip, error) {
	ipv4, err := fetchIpByUrl(ipv4Url)
	if err != nil {
		return nil, fmt.Errorf("error fetching V4 address: %w", err)
	}

	ipv6, err := fetchIpByUrl(ipv6Url)
	if err != nil {
		return nil, fmt.Errorf("error fetching V6 address: %w", err)
	}

	return &Ip{
		V4: ipv4,
		V6: ipv6,
	}, nil
}
