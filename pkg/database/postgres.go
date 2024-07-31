package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

const getPGVersion = "SELECT version();"
const getTableExists = "SELECT EXISTS(SELECT FROM information_schema.tables WHERE  table_schema = $1 AND table_name = $2) as exists;"

type PgxDB struct {
	Conn *pgxpool.Pool
	log  golog.MyLogger
}

func newPgxConn(dbConnectionString string, maxConnectionsInPool int, log golog.MyLogger) (DB, error) {
	var psql PgxDB
	var parsedConfig *pgx.ConnConfig
	var err error
	parsedConfig, err = pgx.ParseConfig(dbConnectionString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error doing pgx.ParseConfig(%s). err: %s", dbConnectionString, err))
	}

	dbHost := parsedConfig.Host
	dbPort := parsedConfig.Port
	dbUser := parsedConfig.User
	dbPass := parsedConfig.Password
	dbName := parsedConfig.Database

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s pool_max_conns=%d", dbHost, dbPort, dbUser, dbPass, dbName, maxConnectionsInPool)

	connPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error("FAILED to connect to database %s with user %s", dbName, dbUser)
		return nil, errors.New(fmt.Sprintf("error connecting to database. err : %s", err))
	} else {
		log.Info("SUCCESS connecting to database %s with user %s", dbName, dbUser)
		// let's first check that we can really make a query by querying the postgres version
		var version string
		errPing := connPool.QueryRow(context.Background(), getPGVersion).Scan(&version)
		if errPing != nil {
			log.Info("something very weird is occurring here... this db connection is probably invalid ! ")
			log.Fatal("got db error retrieving postgres version with : [%s] error: %s", getPGVersion, errPing)
			return nil, errPing
		}

		log.Info("Postgres version: [%s]'", version)
	}

	psql.Conn = connPool
	psql.log = log
	return &psql, err
}

// ExecActionQuery is a postgres helper function for an action query, returning the numbers of rows affected
func (db *PgxDB) ExecActionQuery(sql string, arguments ...interface{}) (rowsAffected int, err error) {
	commandTag, err := db.Conn.Exec(context.Background(), sql, arguments...)
	if err != nil {
		db.log.Error("ExecActionQuery unexpectedly failed with sql: %v . Args(%+v), error : %v", sql, arguments, err)
		return 0, err
	}
	return int(commandTag.RowsAffected()), err
}

func (db *PgxDB) Insert(sql string, arguments ...interface{}) (lastInsertId int, err error) {
	sql4PGX := fmt.Sprintf("%s RETURNING id;", sql)
	err = db.Conn.QueryRow(context.Background(), sql4PGX, arguments...).Scan(&lastInsertId)
	if err != nil {
		db.log.Error(" Insert unexpectedly failed with %v: (%v), error : %v", sql, arguments, err)
		return 0, err
	}
	return lastInsertId, err
}

// GetQueryInt is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryInt(sql string, arguments ...interface{}) (result int, err error) {
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error(" GetQueryInt(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return 0, err
	}
	return result, err
}

// GetQueryBool is a postgres helper function for a query expecting an integer result
func (db *PgxDB) GetQueryBool(sql string, arguments ...interface{}) (result bool, err error) {
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error(" GetQueryBool(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return false, err
	}
	return result, err
}

func (db *PgxDB) GetQueryString(sql string, arguments ...interface{}) (result string, err error) {
	var mayBeResultIsNull *string
	err = db.Conn.QueryRow(context.Background(), sql, arguments...).Scan(&mayBeResultIsNull)
	if err != nil {
		db.log.Error(" GetQueryString(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Error(" GetQueryString() queryRow returned no results with sql: %v ; parameters:(%v)\n", sql, arguments)
		return "", ErrNoRecordFound
	}
	result = *mayBeResultIsNull
	return result, err
}

func (db *PgxDB) GetVersion() (result string, err error) {
	var mayBeResultIsNull *string
	err = db.Conn.QueryRow(context.Background(), getPGVersion).Scan(&mayBeResultIsNull)
	if err != nil {
		db.log.Error(" GetVersion() queryRow unexpectedly failed. error : %v\n", err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Error("GetVersion() queryRow returned no results \n")
		return "", ErrNoRecordFound
	}
	result = *mayBeResultIsNull
	return result, err
}

func (db *PgxDB) GetPGConn() (Conn *pgxpool.Pool, err error) {
	dbVersion, err := db.GetVersion()
	if err != nil || len(dbVersion) < 2 {
		return nil, errors.New("NOT CONNECTED TO DB")
	}
	return db.Conn, nil
}

func (db *PgxDB) DoesTableExist(schema, table string) (exist bool) {
	tableExists, err := db.GetQueryBool(getTableExists, schema, table)
	if err != nil {
		db.log.Error(" DoesTableExist() GetQueryBool returned error:%v \n", err)
		return false
	}
	return tableExists
}

// Close is a postgres helper function to close the connection to the database
func (db *PgxDB) Close() {
	db.Conn.Close()
	return
}
