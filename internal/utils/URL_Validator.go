package utils

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

func isValidURL(raw string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, errors.New("invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("unsupported scheme")
	}

	if parsed.Host == "" {
		return nil, errors.New("missing host")
	}

	return parsed, nil
}

func isSafeHost(host string) error {
	hostname := strings.Split(host, ":")[0] // strip port if present

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return errors.New("cannot resolve host")
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return errors.New("IP address is unsafe (loopback/private)")
		}
	}

	return nil
}

var fastURLRegex = regexp.MustCompile(`^(https?://)([a-zA-Z0-9\-_]+\.)+[a-zA-Z]{2,}(:\d+)?(/.*)?$`)

func isFastValidURL(raw string) error {
	if !fastURLRegex.MatchString(raw) {
		return errors.New("URL fails fast regex check")
	}
	return nil
}

func ValidateURL(raw string) error {
	if err := isFastValidURL(raw); err != nil {
		return err
	}

	parsed, err := isValidURL(raw)
	if err != nil {
		return err
	}

	err = isSafeHost(parsed.Host)
	if err != nil {
		return err
	}

	return nil
}
