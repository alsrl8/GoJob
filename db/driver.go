package db

type Driver interface {
	Close() error
	InsertData(tableName string, data map[string]interface{}) error
}
