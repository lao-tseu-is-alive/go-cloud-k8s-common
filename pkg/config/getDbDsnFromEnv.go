package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

// GetPgDbDsnUrlFromEnvOrPanic returns a valid DSN connection string based on the values of environment variables :
//
//		DB_HOST : string containing a valid Ip Address to use for DB connection
//		DB_PORT : int value between 1 and 65535
//		DB_NAME : string containing the database name
//		DB_USER : string containing the database username
//		DB_PASSWORD : string containing the database user password
//	 DB_SSL_MODE : string containing ssl mode (disable|allow|prefer|require|verify-ca|verify-full)
//	 more info on libpg SSL : https://www.postgresql.org/docs/11/libpq-ssl.html#LIBPQ-SSL-PROTECTION
//	 more info : https://pkg.go.dev/github.com/jackc/pgconn#ParseConfig
func GetPgDbDsnUrlFromEnvOrPanic(defaultIP string, defaultPort int,
	defaultDbName string, defaultDbUser string, defaultSSL string) string {
	srvIP := defaultIP
	srvPort := defaultPort
	dbName := defaultDbName
	dbUser := defaultDbUser
	dbPassword := ""
	dbSslMode := defaultSSL

	var err error
	val, exist := os.LookupEnv("DB_PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			panic(fmt.Errorf("💥💥 ERROR: CONFIG ENV DB_PORT should contain a valid integer. %v", err))
		}
		if srvPort < 1 || srvPort > 65535 {
			panic(fmt.Errorf("💥💥 ERROR: DB_PORT should contain an integer between 1 and 65535. Err: %v", err))
		}
	}
	val, exist = os.LookupEnv("DB_HOST")
	if exist {
		srvIP = val
		if net.ParseIP(srvIP) == nil {
			panic("💥💥 ERROR: CONFIG ENV DB_HOST should contain a valid IP. ")
		}
	}
	val, exist = os.LookupEnv("DB_NAME")
	if exist {
		dbName = val
	}
	val, exist = os.LookupEnv("DB_USER")
	if exist {
		dbUser = val
	}
	val, exist = os.LookupEnv("DB_PASSWORD")
	if exist {
		dbPassword = val
	} else {
		panic("💥💥 ERROR: CONFIG ENV DB_PASSWORD should contain a password value.")
	}
	val, exist = os.LookupEnv("DB_SSL_MODE")
	if exist {
		dbSslMode = val
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbUser, dbPassword, srvIP, srvPort, dbName, dbSslMode)
}
