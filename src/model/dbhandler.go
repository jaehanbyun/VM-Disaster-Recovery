package model

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jaehanbyun/VM-Disaster-Recovery/data"
	_ "github.com/lib/pq"
)

type postgresHandler struct {
	db *sql.DB
}

func (p *postgresHandler) Close() {
	p.db.Close()
}

func (p *postgresHandler) GetWeight() (data.Weight, error) {
	row := p.db.QueryRow("SELECT language, database, webserver FROM weight LIMIT 1")
	var weight data.Weight
	err := row.Scan(&weight.Language, &weight.Database, &weight.Webserver)
	if err != nil {
		return weight, err
	}
	return weight, nil
}

func (p *postgresHandler) SetWeight(weight data.Weight) error {
	_, err := p.db.Exec("INSERT INTO weight (language, database, webserver, threshold) VALUES ($1 $2 $3 $4) ON CONFLICT DO UPDATE",
		weight.Language, weight.Database, weight.Webserver, weight.Threshold)
	return err
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
