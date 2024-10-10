package database

import (
	"database/sql"
	"fmt"
	"log"
)

func NewPostgresConnection(host, port, user, password, dbname string) (*sql.DB, error) {
	psqlinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require TimeZone=UTC",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlinfo) //opens a database connection using the PostgreSQL driver.
	if err != nil {
		return nil, err
	}

	err = db.Ping() //verify that the connection to the database is successful.
	if err != nil {
		return nil, err
	}

	// Set the timezone to UTC for this connection
	_, err = db.Exec("SET TIME ZONE 'UTC'")
	if err != nil {
		log.Printf("Warning: Failed to set timezone to UTC: %v", err)
	}

	// Verify the timezone setting
	var tz string
	err = db.QueryRow("SHOW TIME ZONE").Scan(&tz)
	if err != nil {
		log.Printf("Warning: Failed to verify timezone: %v", err)
	} else if tz != "UTC" {
		log.Printf("Warning: Database timezone is not UTC, it is: %s", tz)
	}

	return db, nil
}
