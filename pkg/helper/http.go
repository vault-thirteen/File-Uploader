package helper

import (
	"net"
	"net/http"
	"strings"

	"github.com/vault-thirteen/auxie/errors"
)

const (
	UrlPathSeparator = "/"
)

const (
	Err_InvalidHeader_XForwardedFor = "invalid header: X-Forwarded-For"
	Err_InvalidHeader_XRealIp       = "invalid header: X-Real-IP"
)

const (
	HttpHeader_XForwardedFor = "X-Forwarded-For"
	HttpHeader_XRealIp       = "X-Real-IP"
)

func GetClientIPAddress(req *http.Request) (ipa string, err error) {
	// X-Forwarded-For.
	xff := req.Header.Get(HttpHeader_XForwardedFor)
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) < 1 {
			return "", errors.ErrorFromText(Err_InvalidHeader_XForwardedFor)
		}

		ipa = strings.TrimSpace(parts[0])

		if len(ipa) == 0 {
			return "", errors.ErrorFromText(Err_InvalidHeader_XForwardedFor)
		}

		return ipa, nil
	}

	// X-Real-IP.
	xrip := req.Header.Get(HttpHeader_XRealIp)
	if xrip != "" {
		xrip = strings.TrimSpace(xrip)

		if len(xrip) == 0 {
			return "", errors.ErrorFromText(Err_InvalidHeader_XRealIp)
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
