package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	// to using mysql driver.
	_ "github.com/go-sql-driver/mysql"
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

	var connectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, name)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}

	dbs[key] = db

	mutex.Unlock()

	return db
}
