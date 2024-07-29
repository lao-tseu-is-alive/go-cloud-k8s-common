package config

import (
	"fmt"
	"os"
	"strconv"
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
