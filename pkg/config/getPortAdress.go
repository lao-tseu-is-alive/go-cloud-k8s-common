package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// GetPortFromEnvOrPanic returns a valid TCP/IP listening port based on the values of environment variable :
//
//	PORT : int value between 1 and 65535 (the parameter defaultPort will be used if env is not defined)
//	 in case the ENV variable PORT exists and contains an invalid integer the functions panics
func GetPortFromEnvOrPanic(defaultPort int) int {
	srvPort := defaultPort
	var err error
	val, exist := os.LookupEnv("PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			panic(fmt.Errorf("💥💥 ERROR: CONFIG ENV PORT should contain a valid integer. %v", err))
		}
	}
	if srvPort < 1 || srvPort > 65535 {
		panic(fmt.Errorf("💥💥 ERROR: PORT should contain an integer between 1 and 65535. Err: %v", err))
	}
	return srvPort
}

// GetListenIpFromEnvOrPanic returns a valid TCP/IP listening address based on the values of environment variable :
//
//	SRV_IP : int value between 1 and 65535 (the parameter defaultPort will be used if env is not defined)
//	 in case the ENV variable PORT exists and contains an invalid integer the functions panics
func GetListenIpFromEnvOrPanic(defaultSrvIp string) string {
	srvIp := defaultSrvIp
	val, exist := os.LookupEnv("SRV_IP")
	if exist {
		srvIp = val
	}
	if net.ParseIP(srvIp) == nil {
		panic("💥💥 ERROR: CONFIG ENV SRV_IP should contain a valid IP. ")
	}
	return srvIp
}

// GetAllowedIpsFromEnvOrPanic returns a list of valid TCP/IP addresses based on the values of env variable ALLOWED_IP
//
//	ALLOWED_IP : comma separated list of valid IP addresses
//	 in case the ENV variable ALLOWED_IP exists and contains invalid IP addresses the functions panics
//	if the ENV variable ALLOWED_IP does not exist the function returns the defaultAllowedIps or panic if invalid
func GetAllowedIpsFromEnvOrPanic(defaultAllowedIps []string) []string {
	allowedIps := defaultAllowedIps
	envValue, exists := os.LookupEnv("ALLOWED_IP")
	if exists {
		allowedIps = []string{}
		ips := strings.Split(envValue, ",")
		for _, ip := range ips {
			trimmedIP := strings.TrimSpace(ip)
			allowedIps = append(allowedIps, trimmedIP)
		}
	}
	for _, ip := range allowedIps {
		if net.ParseIP(ip) == nil {
			panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ALLOWED_IP should contain only valid IP : %s\n", ip))
		}
	}
	if len(allowedIps) > 0 {
		return allowedIps
	}
	panic("💥💥 ERROR: CONFIG ENV ALLOWED_IP should contain at least one valid IP.")
}

// GetAllowedHostsFromEnvOrPanic returns a list of valid TCP/IP addresses based on the values of env variable ALLOWED_HOSTS
//
//	ALLOWED_HOSTS : comma separated list of valid IP addresses
//	in case the ENV variable ALLOWED_HOSTS exists and contains invalid Host addresses the functions panics
func GetAllowedHostsFromEnvOrPanic() []string {
	var allowedHosts []string
	envValue, exist := os.LookupEnv("ALLOWED_HOSTS")
	if !exist {
		panic("💥💥 ERROR: ENV ALLOWED_HOSTS should contain your allowed hosts.")
	}
	if exist {
		allowedHosts = []string{}
		allHosts := strings.Split(envValue, ",")
		for _, hostName := range allHosts {
			trimmedHost := strings.TrimSpace(hostName)
			if trimmedHost == "" {
				continue
			}
			allowedHosts = append(allowedHosts, trimmedHost)
		}
	}
	if len(allowedHosts) > 0 {
		return allowedHosts
	}
	panic("💥💥 ERROR: CONFIG ENV ALLOWED_HOSTS should contain at least one valid Host.")
}
