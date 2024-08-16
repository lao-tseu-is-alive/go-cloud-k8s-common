package metadata

import (
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"strings"
)

type Service struct {
	Log golog.MyLogger
	Db  database.DB
}

const (
	defaultSchema           = "public"
	MetaTableName           = "go_metadata_db_schema"
	CountMetaSQL            = "SELECT COUNT(*) as num FROM go_metadata_db_schema;"
	countMetaServiceSQL     = "SELECT COUNT(*) as num FROM go_metadata_db_schema WHERE service = $1;"
	getVersionServiceSQL    = "SELECT  version FROM go_metadata_db_schema WHERE service = $1"
	updateVersionServiceSQL = "UPDATE public.go_metadata_db_schema SET version = $2 WHERE service = $1"
	selectMetaSQL           = "SELECT  id, service, schema, table_name, version FROM go_metadata_db_schema WHERE service = $1"
	insertMetaSQL           = "INSERT INTO go_metadata_db_schema (service, schema, table_name, version) VALUES ($1,$2,$3,$4)"
	CreateMetaTable         = `
CREATE TABLE IF NOT EXISTS go_metadata_db_schema
(
    id          serial    CONSTRAINT go_metadata_db_schema_pk   primary key,
    service     text                             not null,
    schema      text      default 'public'::text not null,
    table_name  text                             not null,
    version     text                             not null,
    create_time timestamp default now()          not null,
    CONSTRAINT go_metadata_db_schema_unique_service_schema_table
        unique (service, schema, table_name)
);
comment on table go_metadata_db_schema is 'to track version of schema of different micro services';
`
	unableToCountService = "💥💥 unable to count for service %s in the table «metadata». error:%v"
	unableToGetVersion   = "💥💥 unable to retrieve version for service %s in the table «metadata». error:%v"
	unableToSetVersion   = "💥💥 unable to set version for service %s in the table «metadata». error:%v"
)

// CreateMetadataTableOrFail create the metadata table if it is not already present in the db
func (s *Service) CreateMetadataTableOrFail() {
	if s.Db.DoesTableExist(defaultSchema, MetaTableName) {
		numberOfServicesSchema, errMetaTable := s.Db.GetQueryInt(CountMetaSQL)
		if errMetaTable != nil {
			s.Log.Warn("problem counting the rows in metadata table : %v", errMetaTable)
		}
		if numberOfServicesSchema > 0 {
			s.Log.Info("database contains %d service(s) in metadata", numberOfServicesSchema)
		} else {
			s.Log.Warn("database does not contain any  registered service in metadata table")
		}
	} else {
		s.Log.Warn("database does not contain the metadata table, will try to create it...")

		RowsAffected, err := s.Db.ExecActionQuery(CreateMetaTable)
		if err != nil {
			s.Log.Error("problem creating the metadata table : %v", err)
			panic(fmt.Errorf("💥💥 unable to create the table «metadata». error:%v", err))
		}
		s.Log.Info("metadata table was created, rows affected : %v", int(RowsAffected))
	}
}

// GetServiceVersionOrFail allows to retrieve the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) GetServiceVersionOrFail(serviceName string) (found bool, version string) {
	s.Log.Debug("entering GetServiceVersion(%s)", serviceName)
	count, err := s.Db.GetQueryInt(countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Error("GetServiceVersion(%s) could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
		panic(fmt.Errorf(unableToCountService, serviceName, err))
	}
	if count > 0 {
		s.Log.Info("GetServiceVersion(%s) service does exist", serviceName)
		version, err := s.Db.GetQueryString(getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Error("GetServiceVersion(%s) version could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
			panic(fmt.Errorf(unableToGetVersion, serviceName, err))
		}
		return true, version
	} else {
		s.Log.Info("GetServiceVersion(%s) service does not exist", serviceName)
		return false, ""
	}
}

// SetServiceVersionOrFail allows to insert/update the version of the microservice registered (if any) in the metadata table of the db
func (s *Service) SetServiceVersionOrFail(serviceName, version string) {
	s.Log.Debug("entering SetServiceVersion(%s, %s)", serviceName, version)
	count, err := s.Db.GetQueryInt(countMetaServiceSQL, serviceName)
	if err != nil {
		s.Log.Error("SetServiceVersion(%s) could not be retrieved from DB. failed db.Query err: %v", serviceName, err)
		panic(fmt.Errorf(unableToCountService, serviceName, err))
	}

	if count > 0 {
		s.Log.Info("GetServiceVersion(%s) service does exist", serviceName)
		versionInDB, err := s.Db.GetQueryString(getVersionServiceSQL, serviceName)
		if err != nil {
			s.Log.Error("SetServiceVersion(%s) version could not be retrieved from DB. failed db.Query err: %v", serviceName, err)

			panic(fmt.Errorf(unableToGetVersion, serviceName, err))
		}
		if versionInDB == version {
			s.Log.Info("SetServiceVersion(%s) service does already exist with this version, nothing to do", serviceName)
		} else {
			rowsAffected, err := s.Db.ExecActionQuery(updateVersionServiceSQL, serviceName, version)
			if err != nil {
				s.Log.Error("SetServiceVersion(%s) version could not be updated in DB. failed db.Query err: %v", serviceName, err)
				panic(fmt.Errorf(unableToSetVersion, serviceName, err))
			}
			s.Log.Info("SetServiceVersion(%s) service updated to version: %s (%d rows)", serviceName, version, rowsAffected)
		}
	} else {
		s.Log.Info("SetServiceVersion(%s) service does not exist will insert it", serviceName)
		rowsAffected, err := s.Db.ExecActionQuery(insertMetaSQL, serviceName, defaultSchema, strings.ToLower(serviceName), version)
		if err != nil {
			s.Log.Error("SetServiceVersion(%s) version could not be inserted in DB. failed db.Query err: %v", serviceName, err)
			panic(fmt.Errorf(unableToSetVersion, serviceName, err))
		}
		s.Log.Info("SetServiceVersion(%s) service inserted with version: %s (%d rows)", serviceName, version, rowsAffected)
	}
}
