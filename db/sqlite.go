package db

import (
	"GoJob/xlog"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"sync"
)

type Sqlite struct {
	*sql.DB
}

var sqlite *Sqlite
var onceSqlite sync.Once

func NewSqlite() *Sqlite {
	onceSqlite.Do(func() {
		sqlite = &Sqlite{}
		err := sqlite.connect()
		if err != nil {
			xlog.Logger.Error(err)
			return
		}
	})
	return sqlite
}

func (s *Sqlite) connect() error {
	db, err := sql.Open("sqlite3", "./gojob.db")
	if err != nil {
		xlog.Logger.Error(err)
		return errors.New("failed to connect database")
	}
	s.DB = db

	err = s.initTable()
	if err != nil {
		xlog.Logger.Error(err)
		return errors.New("failed to connect database")
	}

	return nil
}

func (s *Sqlite) initTable() error {
	_, err := s.DB.Exec("CREATE TABLE IF NOT EXISTS jumpit (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, company TEXT,  skills TEXT, link TEXT)")
	if err != nil {
		xlog.Logger.Error(err)
		return errors.New(fmt.Sprintf("failed to initialize sqlite table: %s", "jumpit"))
	}

	return nil
}

func (s *Sqlite) Close() error {
	err := s.DB.Close()
	if err != nil {
		xlog.Logger.Error(err)
		return err
	}
	panic("sqlite3 close")
}

func (s *Sqlite) InsertData(tableName string, data map[string]interface{}) error {
	var columns []string
	var placeholders []string
	var values []interface{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := s.Exec(query, values...)
	if err != nil {
		xlog.Logger.Error(err)
		return err
	}
	return nil
}

func (s *Sqlite) SelectData(tableName string, where string, args ...interface{}) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s %s",
		tableName,
		where,
	)

	data, err := s.Query(query, args...)
	if err != nil {
		xlog.Logger.Error(err)
		return []map[string]interface{}{}, err
	}
	defer func(data *sql.Rows) {
		err = data.Close()
		if err != nil {
			xlog.Logger.Error(err)
			return
		}
	}(data)

	var results []map[string]interface{}
	columns, err := data.Columns()
	if err != nil {
		xlog.Logger.Error(err)
		return nil, err
	}

	for data.Next() {
		row := make(map[string]interface{})
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			columnPointers[i] = new(interface{})
		}
		err := data.Scan(columnPointers...)
		if err != nil {
			xlog.Logger.Error(err)
			return nil, err
		}
		for i, colName := range columns {
			row[colName] = *(columnPointers[i].(*interface{}))
		}
		results = append(results, row)
	}
	if err = data.Err(); err != nil {
		xlog.Logger.Error(err)
		return nil, err
	}
	return results, nil
}

func TestSqlite() {
	db, err := sql.Open("sqlite3", "./gojob.db")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			xlog.Logger.Error(err)
		}
	}(db)

	// Create table
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER)")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}
	_, err = stmt.Exec()
	if err != nil {
		return
	}

	// Insert dummy data
	stmt, err = db.Prepare("INSERT INTO user (name, age) VALUES (?, ?)")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	_, err = stmt.Exec("John Doe", 20)
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	// Query the database
	rows, err := db.Query("SELECT * FROM user")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			xlog.Logger.Error(err)
			return
		}
	}(rows)

	for rows.Next() {
		var id int
		var name string
		var age int
		err = rows.Scan(&id, &name, &age)
		if err != nil {
			xlog.Logger.Error(err)
			return
		}

		xlog.Logger.Info(id, name, age)
	}
}
