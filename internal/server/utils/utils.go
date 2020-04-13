package utils

import "net"

func ParseHost(fullHost string) string {
	host, _, err := net.SplitHostPort(fullHost)
	if err != nil { // no port
		return fullHost
	}
	return host
}
