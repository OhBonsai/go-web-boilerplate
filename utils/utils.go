package utils

import (
	"github.com/OhBonsai/go-web-boilerplate/model"
	"strings"
	"net"
	"net/http"
)

func GetIpAddress(r *http.Request) string {
	address := ""

	header := r.Header.Get(model.HEADER_FORWARDED)
	if len(header) > 0 {
		addresses := strings.Fields(header)
		if len(addresses) > 0 {
			address = strings.TrimRight(addresses[0], ",")
		}
	}

	if len(address) == 0 {
		address = r.Header.Get(model.HEADER_REAL_IP)
	}

	if len(address) == 0 {
		address, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return address
}
