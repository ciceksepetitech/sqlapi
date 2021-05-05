package mssql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	// to using mssql driver.
	_ "github.com/denisenkom/go-mssqldb"
)

var (
	dbs   map[string]*sql.DB
	mutex *sync.Mutex
)

func init() {
	dbs = map[string]*sql.DB{}
	mutex = &sync.Mutex{}
}

// I to get instance of database object.
func I(user string, password string, host string, name string) *sql.DB {
	key := fmt.Sprintf("%s-%s", host, name)

	if val, ok := dbs[key]; ok {
		return val
	}

	mutex.Lock()

	if val, ok := dbs[key]; ok {
		return val
	}

	port := "1433"
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
		port = parts[1]
	}

	db, err := sql.Open("mssql", fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;port=%s", host, user, password, name, port))
	if err != nil {
		panic(err)
	}

	dbs[key] = db

	mutex.Unlock()

	return db
}
