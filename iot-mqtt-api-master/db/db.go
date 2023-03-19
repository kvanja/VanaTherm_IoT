package db

import (
	"database/sql"
	"fmt"
	"iot-mqtt-api/config"

	_ "github.com/denisenkom/go-mssqldb"
)

// Database handles DB connections
type Database struct {
	Conn *sql.DB
}

// Connect create new Database struct and connects to DB
func Connect(dbConfig config.Database) (*Database, error) {
	conn, err := sql.Open("mssql", databaseURL(dbConfig))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &Database{Conn: conn}, nil
}

// Disconnect terminates connection to DB pool
func (db *Database) Disconnect() {
	db.Conn.Close()
}

// DatabaseURL returns database url in DSN format:
// postgres://USER:PASS@HOSTNAME:PORT/DBNAME
func databaseURL(c config.Database) string {
	return fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
	)
}

func (db *Database) GetDatabaseURL(c config.Database) string {
	return fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
	)
}
