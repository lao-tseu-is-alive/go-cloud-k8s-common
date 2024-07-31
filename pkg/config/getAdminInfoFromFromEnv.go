package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const minUserNameLength = 5
const minUserEmailLength = 12
const minUserPasswordLength = 8

// GetAdminUserFromFromEnvOrPanic returns the admin user to be used with JWT authentication from the content of the env variable :
// ADMIN_USER : string containing the username to use for the administrative account
func GetAdminUserFromFromEnvOrPanic(defaultAdminUser string) string {
	adminUser := defaultAdminUser
	val, exist := os.LookupEnv("ADMIN_USER")
	if exist {
		adminUser = val
	}
	if utf8.RuneCountInString(adminUser) < minUserNameLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_USER should contain at least %d characters (got %d).",
			minUserNameLength, utf8.RuneCountInString(val)))
	}
	return fmt.Sprintf("%s", adminUser)
}

// GetAdminEmailFromFromEnvOrPanic returns the admin user to be used with JWT authentication from the content of the env variable :
// ADMIN_EMAIL : string containing the username to use for the administrative account
func GetAdminEmailFromFromEnvOrPanic(defaultAdminEmail string) string {
	adminEmail := defaultAdminEmail
	val, exist := os.LookupEnv("ADMIN_EMAIL")
	if exist {
		adminEmail = val
	}
	if utf8.RuneCountInString(adminEmail) < minUserEmailLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_EMAIL should contain at least %d characters (got %d).",
			minUserEmailLength, utf8.RuneCountInString(val)))
	}
	_, err := mail.ParseAddress(adminEmail)
	if err != nil {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_EMAIL should be a valid email address."))
	}
	hasSpecial := false
	specialRune := '@'
	for _, c := range adminEmail {
		switch {
		case c == '@' || c == '.' || c == '_' || c == '-':
			break
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
			specialRune = c
		}
	}
	if hasSpecial {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_EMAIL should not contain special character except @ - _.'%v' is not valid", specialRune))
	}
	return fmt.Sprintf("%s", adminEmail)
}

// GetAdminIdFromFromEnvOrPanic returns the admin user id to be used with JWT authentication from the content of the env variable
// ADMIN_ID : string containing the user id to use for the administrative account
func GetAdminIdFromFromEnvOrPanic(defaultAdminId int) int {
	adminId := defaultAdminId
	var err error
	val, exist := os.LookupEnv("ADMIN_ID")
	if exist {
		adminId, err = strconv.Atoi(val)
		if err != nil {
			panic(fmt.Errorf("💥💥 ERROR: ENV ADMIN_ID should contain a valid integer. %v", err))
		}
	}
	return adminId
}

// GetAdminPasswordFromFromEnvOrPanic returns the admin password to be used with JWT authentication from the content of the env variable :
//
//	ADMIN_PASSWORD : string containing the password to use for the administrative account
func GetAdminPasswordFromFromEnvOrPanic() string {
	adminPassword := ""
	val, exist :=
		os.LookupEnv("ADMIN_PASSWORD")
	if !exist {
		panic("💥💥 ERROR: ENV ADMIN_PASSWORD should contain your JWT secret.")
	}
	adminPassword = val
	if utf8.RuneCountInString(adminPassword) < minUserPasswordLength {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_PASSWORD should contain at least %d characters (got %d).",
			minUserPasswordLength, utf8.RuneCountInString(val)))
	}
	if !VerifyPasswordComplexity(adminPassword) {
		panic(fmt.Sprintf("💥💥 ERROR: CONFIG ENV ADMIN_PASSWORD should contain at least one lowercase letter, one uppercase letter, one digit and one special	character. No white space, #, or | or ' character in it."))
	}
	return fmt.Sprintf("%s", adminPassword)
}

// VerifyPasswordComplexity checks if the password meets the minimum requirements of complexity
// At least one lowercase letter,one uppercase letter, one digit and one special character
// No white space, #, or | or ' character in it
func VerifyPasswordComplexity(s string) bool {
	var hasNumber, hasUpperCase, hasLowercase, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpperCase = true
		case unicode.IsLower(c):
			hasLowercase = true
		case c == '#' || c == '|' || c == '\'' || unicode.IsSpace(c):
			return false
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasNumber && hasUpperCase && hasLowercase && hasSpecial
}
