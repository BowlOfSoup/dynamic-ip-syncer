package icanhazip

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const ipv4Url = "https://ipv4.icanhazip.com/"
const ipv6Url = "https://ipv6.icanhazip.com/"

type Ip struct {
	V4 string
	V6 string
}

type ipFamily int

const (
	ipFamilyV4 ipFamily = iota
	ipFamilyV6
)

func fetchIpByUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IP from %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			log.Fatalf("[external request] error: failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from %s: %w", url, err)
	}

	return strings.TrimSpace(string(body)), nil
}

func (f ipFamily) label() string {
	switch f {
	case ipFamilyV4:
		return "V4"
	case ipFamilyV6:
		return "V6"
	default:
		return "unknown"
	}
}

func validateIP(ip string, family ipFamily, source string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid %s IP from %s: %s", family.label(), source, ip)
	}

	isV4 := parsed.To4() != nil
	if family == ipFamilyV4 && !isV4 {
		return fmt.Errorf("%s did not return an IPv4 address: %s", source, ip)
	}
	if family == ipFamilyV6 && isV4 {
		return fmt.Errorf("%s did not return an IPv6 address: %s", source, ip)
	}

	return nil
}

func fetchValidatedIP(url string, family ipFamily) (string, error) {
	sourceIP, err := fetchIpByUrl(url)
	if err != nil {
		return "", fmt.Errorf("error fetching %s IP from %s: %w", family.label(), url, err)
	}

	if err := validateIP(sourceIP, family, url); err != nil {
		return "", err
	}

	return sourceIP, nil
}

func resolvedURL(sourceURL *url.URL, fallback string) string {
	if sourceURL == nil {
		return fallback
	}

	return sourceURL.String()
}

func GetIpAddresses(sourceUrlV4 *url.URL, sourceUrlV6 *url.URL) (*Ip, error) {
	ipv4URL := resolvedURL(sourceUrlV4, ipv4Url)
	ipv6URL := resolvedURL(sourceUrlV6, ipv6Url)

	ipv4, err := fetchValidatedIP(ipv4URL, ipFamilyV4)
	if err != nil {
		return nil, err
	}

	ipv6, err := fetchValidatedIP(ipv6URL, ipFamilyV6)
	if err != nil {
		return nil, err
	}

	return &Ip{
		V4: ipv4,
		V6: ipv6,
	}, nil
}
