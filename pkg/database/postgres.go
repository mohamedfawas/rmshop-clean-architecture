package database

import (
	"database/sql"
	"fmt"
)

func NewPostgresConnection(host, port, user, password, dbname string) (*sql.DB, error) {
	psqlinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlinfo) //opens a database connection using the PostgreSQL driver.
	if err != nil {
		return nil, err
	}

	err = db.Ping() //verify that the connection to the database is successful.
	if err != nil {
		return nil, err
	}

	return db, nil
}
