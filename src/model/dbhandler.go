package model

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type postgresHandler struct {
	db *sql.DB
}

func (p *postgresHandler) Close() {
	p.db.Close()
}

func newPostgresHandler() DBHandler {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		"localhost", 5432, "postgres", "postgres", "postgres",
	)

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = database.Ping()
	if err != nil {
		panic(err)
	}

	createWeight, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS weight (
				laugage NUMERIC,
				database NUMERIC,
				webserver NUMERIC,
				threshold NUMERIC
			);`)
	createWeight.Exec()

	createVMInfo, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS vminfo (
			id STRING PRIMARY KEY,
			os STRING,
			lauguage JSON,
			database JSON,
			webserver JSON
		);`)
	createVMInfo.Exec()

	return &postgresHandler{database}
}
