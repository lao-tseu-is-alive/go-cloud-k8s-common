package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetPortFromEnv returns a valid TCP/IP listening ':PORT' string based on the values of environment variable :
//
//		PORT : int value between 1 and 65535 (the parameter defaultPort will be used if env is not defined)
//	 in case the ENV variable PORT exists and contains an invalid integer the functions returns an empty string and an error
func GetPortFromEnv(defaultPort int) (string, error) {
	srvPort := defaultPort

	var err error
	val, exist := os.LookupEnv("PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("ERROR: CONFIG ENV PORT should contain a valid integer. %v", err)
		}
		if srvPort < 1 || srvPort > 65535 {
			return "", fmt.Errorf("ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535. Err: %v", err)
		}
	}
	return fmt.Sprintf(":%d", srvPort), nil
}
