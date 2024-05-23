package util

import (
	"github.com/slink-go/logging"
	"net"
	"strconv"
	"strings"
)

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func ParseEndpoint(input string) (string, string, int) {
	scheme := "http"
	host := ""
	port := 0
	switch {
	case strings.HasPrefix(input, "https://"):
		scheme = "https"
		host, port = doParseEndpoint(input, "https://")
	case strings.HasPrefix(input, "http://"):
		scheme = "http"
		host, port = doParseEndpoint(input, "http://")
	case strings.HasPrefix(input, "grpcs://"):
		scheme = "grpcs"
		host, port = doParseEndpoint(input, "grpcs://")
	case strings.HasPrefix(input, "grpc://"):
		scheme = "grpc"
		host, port = doParseEndpoint(input, "grpc://")
	default:
		host, port = doParseEndpoint(input, "http://")
	}
	return scheme, host, port
}
func doParseEndpoint(input, prefix string) (string, int) {
	suffix := input[len(prefix):]
	parts := strings.Split(suffix, ":")
	if len(parts) < 2 {
		return parts[0], 0
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		logging.GetLogger("util").Warning("could not parse port value from %s: %s", parts[1], err)
	}
	return parts[0], port
}
