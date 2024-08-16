package metadata

import (
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockDB is a mock implementation of the database.DB interface
type MockDB struct {
	TableExists       bool
	QueryIntResult    int
	QueryIntErr       error
	QueryStringResult string
	QueryStringErr    error
	ExecActionRows    int64
	ExecActionErr     error
}

func (m *MockDB) Insert(sql string, arguments ...interface{}) (lastInsertId int, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) GetQueryBool(sql string, arguments ...interface{}) (result bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) GetVersion() (result string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) GetPGConn() (Conn *pgxpool.Pool, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) Close() {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) DoesTableExist(schema, table string) bool {
	return m.TableExists
}

func (m *MockDB) GetQueryInt(query string, args ...interface{}) (int, error) {
	return m.QueryIntResult, m.QueryIntErr
}

func (m *MockDB) GetQueryString(query string, args ...interface{}) (string, error) {
	return m.QueryStringResult, m.QueryStringErr
}

func (m *MockDB) ExecActionQuery(sql string, arguments ...interface{}) (rowsAffected int, err error) {
	return int(m.ExecActionRows), m.ExecActionErr
}

func TestService_CreateMetadataTableOrFail(t *testing.T) {
	// Test case 1: Table exists, count succeeds
	mockDB := &MockDB{
		TableExists:    true,
		QueryIntResult: 5, // Simulate 5 services existing
		QueryIntErr:    nil,
	}
	logger, _ := golog.NewLogger("zap", golog.WarnLevel, "Test_CreateMetadataTableOrFail")
	service := &Service{Log: logger, Db: mockDB}
	service.CreateMetadataTableOrFail()

	// Test case 2: Table doesn't exist, creation succeeds
	mockDB = &MockDB{
		TableExists:    false,
		ExecActionRows: 1, // Simulate 1 row affected by table creation
		ExecActionErr:  nil,
	}
	service = &Service{Log: logger, Db: mockDB}
	service.CreateMetadataTableOrFail()

	// Test case 3: Table doesn't exist, creation fails
	mockDB = &MockDB{
		TableExists:    false,
		ExecActionRows: 0,
		ExecActionErr:  errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.CreateMetadataTableOrFail() }, "Expected panic when table creation fails")
}

func TestService_GetServiceVersionOrFail(t *testing.T) {
	logger, _ := golog.NewLogger("zap", golog.WarnLevel, "Test_GetServiceVersionOrFail")

	// Test case 1: Service exists, version retrieval succeeds
	mockDB := &MockDB{
		QueryIntResult:    1, // Service exists
		QueryIntErr:       nil,
		QueryStringResult: "1.2.3",
		QueryStringErr:    nil,
	}
	service := &Service{Log: logger, Db: mockDB}
	found, version := service.GetServiceVersionOrFail("test-service")
	assert.True(t, found)
	assert.Equal(t, "1.2.3", version)

	// Test case 2: Service doesn't exist
	mockDB = &MockDB{
		QueryIntResult: 0, // Service doesn't exist
		QueryIntErr:    nil,
	}
	service = &Service{Log: logger, Db: mockDB}
	found, version = service.GetServiceVersionOrFail("nonexistent-service")
	assert.False(t, found)
	assert.Equal(t, "", version)

	// Test case 3: Error counting services
	mockDB = &MockDB{
		QueryIntErr: errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.GetServiceVersionOrFail("error-service") }, "Expected panic when counting services fails")

	// Test case 4: Error retrieving version
	mockDB = &MockDB{
		QueryIntResult: 1, // Service exists
		QueryIntErr:    nil,
		QueryStringErr: errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.GetServiceVersionOrFail("error-service") }, "Expected panic when retrieving version fails")
}

func TestService_SetServiceVersionOrFail(t *testing.T) {
	logger, _ := golog.NewLogger("zap", golog.WarnLevel, "TestService_SetServiceVersionOrFail")
	// Test case 1: Service exists, version update succeeds
	mockDB := &MockDB{
		QueryIntResult:    1, // Service exists
		QueryIntErr:       nil,
		QueryStringResult: "1.0.0", // Existing version
		QueryStringErr:    nil,
		ExecActionRows:    1, // Simulate 1 row affected by update
		ExecActionErr:     nil,
	}
	service := &Service{Log: logger, Db: mockDB}
	service.SetServiceVersionOrFail("test-service", "2.0.0")

	// Test case 2: Service doesn't exist, insert succeeds
	mockDB = &MockDB{
		QueryIntResult: 0, // Service doesn't exist
		QueryIntErr:    nil,
		ExecActionRows: 1, // Simulate 1 row affected by insert
		ExecActionErr:  nil,
	}
	service = &Service{Log: logger, Db: mockDB}
	service.SetServiceVersionOrFail("new-service", "1.0.0")

	// Test case 3: Error counting services
	mockDB = &MockDB{
		QueryIntErr: errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.SetServiceVersionOrFail("error-service", "") }, "Expected panic when counting services fails")

	// Test case 4: Error retrieving version
	mockDB = &MockDB{
		QueryIntResult: 1, // Service exists
		QueryIntErr:    nil,
		QueryStringErr: errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.SetServiceVersionOrFail("error-service", "") }, "Expected panic when retrieving version fails")

	// Test case 5: Service exists, version also succeeds
	mockDB = &MockDB{
		QueryIntResult:    1, // Service exists
		QueryIntErr:       nil,
		QueryStringResult: "1.0.0", // Existing version
		QueryStringErr:    nil,
		ExecActionRows:    1, // Simulate 1 row affected by update
		ExecActionErr:     nil,
	}
	service = &Service{Log: logger, Db: mockDB}
	service.SetServiceVersionOrFail("test-service", "1.0.0")

	// Test case 6: Service exists, version update fails
	mockDB = &MockDB{
		QueryIntResult:    1, // Service exists
		QueryIntErr:       nil,
		QueryStringResult: "1.0.0", // Existing version
		QueryStringErr:    nil,
		ExecActionRows:    0, // Simulate 0 rows affected by update
		ExecActionErr:     errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.SetServiceVersionOrFail("test-service", "2.0.0") }, "Expected panic when updating version fails")

	// Test case 7: Service does not exist, insert fails
	mockDB = &MockDB{
		QueryIntResult: 0, // Service doesn't exist
		QueryIntErr:    nil,
		ExecActionRows: 0, // Simulate 0 rows affected by insert
		ExecActionErr:  errors.New("simulated error"),
	}
	service = &Service{Log: logger, Db: mockDB}
	assert.Panics(t, func() { service.SetServiceVersionOrFail("new-service", "1.0.0") }, "Expected panic when inserting version fails")

}
