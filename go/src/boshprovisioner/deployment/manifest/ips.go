package manifest

import (
	gonet "net"
	"strings"

	bosherr "bosh/errors"
)

func NewIPsFromStrings(strs []string) ([]gonet.IP, error) {
	var ips []gonet.IP

	for _, str := range strs {
		ip := gonet.ParseIP(strings.Trim(str, " "))
		if ip == nil {
			return ips, bosherr.New("Parsing IP %s", str)
		}

		ips = append(ips, ip)
	}

	return ips, nil
}
