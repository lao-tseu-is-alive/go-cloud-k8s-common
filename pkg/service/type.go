package service

import (
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

type Service struct {
	Name        string
	Log         golog.MyLogger
	DbConn      database.DB
	JwtSecret   []byte
	JwtDuration int
	adminUser   string
	adminHash   string
}

// UserLogin defines model for UserLogin.
type UserLogin struct {
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
}
