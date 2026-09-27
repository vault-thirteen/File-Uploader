package helper

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

const (
	UrlPathSeparator = "/"
)

func GetClientIPAddress(req *http.Request) (ipa string, err error) {
	// X-Forwarded-For.
	xff := req.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) < 1 {
			return "", errors.New("invalid header: X-Forwarded-For")
		}

		ipa = strings.TrimSpace(parts[0])

		if len(ipa) == 0 {
			return "", errors.New("invalid header: X-Forwarded-For")
		}

		return ipa, nil
	}

	// X-Real-IP.
	xrip := req.Header.Get("X-Real-IP")
	if xrip != "" {
		xrip = strings.TrimSpace(xrip)

		if len(xrip) == 0 {
			return "", errors.New("invalid header: X-Real-IP")
		}

		return xrip, nil
	}

	// Direct IP Address, formatted as "IPA:port".
	var host string
	host, _, err = net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", err
	}

	return host, nil
}
