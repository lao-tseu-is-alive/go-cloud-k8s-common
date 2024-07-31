package gohttp

import (
	"crypto/sha256"
	"fmt"
)

type Authentication interface {
	AuthenticateUser(user, passwordHash string) bool
	GetUserInfoFromLogin(login string) *UserInfo
}

// SimpleAdminAuthenticator Create a struct that will implement the Authentication interface
type SimpleAdminAuthenticator struct {
	// You can add fields here if needed, e.g., a database connection
	mainAdminUserLogin    string
	mainAdminPasswordHash string
	mainAdminEmail        string
	mainAdminId           int
	jwtChecker            JwtChecker
}

// AuthenticateUser Implement the AuthenticateUser method for SimpleAdminAuthenticator
func (sa *SimpleAdminAuthenticator) AuthenticateUser(userLogin, passwordHash string) bool {
	if userLogin == sa.mainAdminUserLogin && passwordHash == sa.mainAdminPasswordHash {
		return true
	}
	sa.jwtChecker.GetLogger().Info("User %s was not authenticated", userLogin)
	return false
}

// GetUserInfoFromLogin Get the JWT claims from the login User
func (sa *SimpleAdminAuthenticator) GetUserInfoFromLogin(login string) *UserInfo {
	user := &UserInfo{
		UserId:    sa.mainAdminId,
		UserName:  fmt.Sprintf("SimpleAdminAuthenticator_%s", sa.mainAdminUserLogin),
		UserEmail: sa.mainAdminEmail,
		UserLogin: login,
		IsAdmin:   true,
	}
	return user
}

// NewSimpleAdminAuthenticator Function to create an instance of SimpleAdminAuthenticator
func NewSimpleAdminAuthenticator(mainAdminUser, mainAdminPassword, mainAdminEmail string, mainAdminId int, jwtCheck JwtChecker) Authentication {
	l := jwtCheck.GetLogger()
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	l.Debug("mainAdminUserLogin: %s\n", mainAdminUser)
	l.Debug("mainAdminPasswordHash: %s\n", mainAdminPasswordHash)
	return &SimpleAdminAuthenticator{
		mainAdminUserLogin:    mainAdminUser,
		mainAdminPasswordHash: mainAdminPasswordHash,
		mainAdminEmail:        mainAdminEmail,
		mainAdminId:           mainAdminId,
		jwtChecker:            jwtCheck,
	}
}
