package gohttp

import (
	"crypto/sha256"
	"fmt"
	"log"
)

type Authentication interface {
	AuthenticateUser(user, passwordHash string) bool
}

// SimpleAdminAuthenticator Create a struct that will implement the Authentication interface
type SimpleAdminAuthenticator struct {
	// You can add fields here if needed, e.g., a database connection
	mainAdminUser         string
	mainAdminPasswordHash string
}

// AuthenticateUser Implement the AuthenticateUser method for SimpleAdminAuthenticator
func (sa *SimpleAdminAuthenticator) AuthenticateUser(user, passwordHash string) bool {
	if user == sa.mainAdminUser && passwordHash == sa.mainAdminPasswordHash {
		return true
	}
	return false
}

// NewSimpleAuthenticator Function to create an instance of SimpleAdminAuthenticator
func NewSimpleAuthenticator() Authentication {
	return &SimpleAdminAuthenticator{}
}

// NewSimpleAdminAuthenticator Function to create an instance of SimpleAdminAuthenticator
func NewSimpleAdminAuthenticator(mainAdminUser, mainAdminPassword string) Authentication {
	h := sha256.New()
	h.Write([]byte(mainAdminPassword))
	mainAdminPasswordHash := fmt.Sprintf("%x", h.Sum(nil))
	log.Printf("mainAdminUser: %s\n", mainAdminUser)
	log.Printf("mainAdminPasswordHash: %s\n", mainAdminPasswordHash)
	return &SimpleAdminAuthenticator{
		mainAdminUser:         mainAdminUser,
		mainAdminPasswordHash: mainAdminPasswordHash,
	}
}
