package mysql

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql" //Es el conector para mysql
)

var (
	once sync.Once
	db   *sql.DB
	err  error
)

//Connect is a function that permited the connection to mysql
func Connect() *sql.DB {
	user := "root"
	password := "system"
	server := "localhost"
	database := "network"

	once.Do(func() {
		db, err = sql.Open("mysql", user+":"+password+"@tcp("+server+")/"+database)
		if err != nil {
			log.Println(err.Error())
		}
	})

	return db
}
