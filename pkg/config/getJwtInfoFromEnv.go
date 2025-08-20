package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"
)

const (
	minSecretLength     = 16
	minContextKeyLength = 6
)

// GetJwtSecretFromEnvOrPanic returns a secret to be used with JWT based on the content of the env variable
// JWT_SECRET : should exist and contain a string with your secret or this function will panic
func GetJwtSecretFromEnvOrPanic() string {
	val, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		panic("💥💥 ERROR: ENV JWT_SECRET should contain your JWT secret.")
	}
	if utf8.RuneCountInString(val) < minSecretLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV JWT_SECRET should contain at least %d characters (got %d).",
			minSecretLength, utf8.RuneCountInString(val)))
	}
	return fmt.Sprintf("%s", val)
}

// GetJwtIssuerFromEnvOrPanic returns a secret to be used with JWT based on the content of the env variable
// JWT_ISSUER_ID : should exist and contain a string with your secret or this function will panic
func GetJwtIssuerFromEnvOrPanic() string {
	val, exist := os.LookupEnv("JWT_ISSUER_ID")
	if !exist {
		panic("💥💥 ERROR: ENV JWT_ISSUER_ID should contain your JWT ISSUER ID secret.")
	}
	if utf8.RuneCountInString(val) < minSecretLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV JWT_ISSUER_ID should contain at least %d characters (got %d).",
			minSecretLength, utf8.RuneCountInString(val)))
	}
	return fmt.Sprintf("%s", val)
}

// GetJwtContextKeyFromEnvOrPanic returns a secret to be used with JWT based on the content of the env variable
// JWT_CONTEXT_KEY : should exist and contain a string with your secret or this function will panic
func GetJwtContextKeyFromEnvOrPanic() string {
	val, exist := os.LookupEnv("JWT_CONTEXT_KEY")
	if !exist {
		panic("💥💥 ERROR: ENV JWT_CONTEXT_KEY should contain your JWT CONTEXT KEY.")
	}
	if utf8.RuneCountInString(val) < minContextKeyLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV JWT_CONTEXT_KEY should contain at least %d characters (got %d).",
			minContextKeyLength, utf8.RuneCountInString(val)))
	}
	// Check if the value contains only letters
	match, _ := regexp.MatchString("^[a-zA-Z]+$", val)
	if !match {
		panic("💥💥 ERROR: CONFIG ENV JWT_CONTEXT_KEY should contain only letters (a-z, A-Z).")
	}
	return fmt.Sprintf("%s", val)
}

// GetJwtAuthUrlFromEnvOrPanic returns the url to be used for JWT authentication based on the content of the env variable
// JWT_AUTH_URL : should exist and contain a string with your url to be used or this function will panic
func GetJwtAuthUrlFromEnvOrPanic() string {
	val, exist := os.LookupEnv("JWT_AUTH_URL")
	if !exist {
		panic("💥💥 ERROR: ENV JWT_AUTH_URL should contain your JWT AUTHENTICATION URL.")
	}
	// Check if the value contains valid url
	match, _ := regexp.MatchString("^(?:(?:https?|ftp):\\/\\/(?:[^@]+@)?[^:\\/?#]+(?::\\d+)?(?:\\/[^?#]*)?|\\/[^?#]*)$", val)
	if !match {
		panic("💥💥 ERROR: CONFIG ENV JWT_AUTH_URL should contain a valid url")
	}
	return fmt.Sprintf("%s", val)
}

// GetJwtDurationFromEnvOrPanic returns a number  string based on the values of environment variable :
// JWT_DURATION_MINUTES : int value between 1 and 1440 minutes, 24H or 1 day is the maximum duration
// the parameter defaultJwtDuration will be used if this env variable is not defined
// in case the ENV variable JWT_DURATION_MINUTES exists and contains an invalid integer the functions ends execution with Fatalreturns 0 and an error
func GetJwtDurationFromEnvOrPanic(defaultJwtDuration int) int {
	JwtDuration := defaultJwtDuration
	var err error
	val, exist := os.LookupEnv("JWT_DURATION_MINUTES")
	if exist {
		JwtDuration, err = strconv.Atoi(val)
		if err != nil {
			panic("💥💥 ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain a valid integer.")
		}
	}
	if JwtDuration < 1 || JwtDuration > 1440 {
		panic("💥💥 ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain an integer between 1 and 1440")
	}
	return JwtDuration
}

// GetJwtCookieNameFromEnv returns a the name of the http-only cookie to be used to use JWT from env variable
// JWT_COOKIE_NAME : should exist and contain a string with your cookie name or this function will use the passed default
func GetJwtCookieNameFromEnv(defaultName string) string {
	val, exist := os.LookupEnv("JWT_COOKIE_NAME")
	if !exist {
		return defaultName
	}
	return fmt.Sprintf("%s", val)
}

// GetJwtStatusUrlFromEnv returns the url to be used to check JWT token from env variable
// JWT_STATUS_URL : should exist and contain a relative url for status token check or this function will use the passed default
func GetJwtStatusUrlFromEnv(defaultName string) string {
	val, exist := os.LookupEnv("JWT_STATUS_URL")
	if !exist {
		return defaultName
	}
	return fmt.Sprintf("%s", val)
}
