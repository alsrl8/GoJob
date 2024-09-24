package db

type Driver interface {
	Close() error
	InsertData(tableName string, data map[string]interface{}) error
	SelectData(tableName string, where string, args ...interface{}) ([]map[string]interface{}, error)
	DeleteData(tableName string, where string, args ...interface{}) error
}
