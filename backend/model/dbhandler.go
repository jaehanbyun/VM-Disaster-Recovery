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
	row := p.db.QueryRow("SELECT language, database, webserver, threshold FROM weight WHERE id = 1")
	var weight data.Weight
	err := row.Scan(&weight.Language, &weight.Database, &weight.Webserver, &weight.Threshold)
	if err != nil {
		return weight, err
	}
	return weight, nil
}

func (p *postgresHandler) SetWeight(weight data.Weight) error {
	_, err := p.db.Exec(`INSERT INTO weight (id, language, database, webserver, threshold) 
                         VALUES (1, $1, $2, $3, $4) 
                         ON CONFLICT (id)
                         DO UPDATE SET language = EXCLUDED.language, 
                                       database = EXCLUDED.database, 
                                       webserver = EXCLUDED.webserver, 
                                       threshold = EXCLUDED.threshold`,
		weight.Language, weight.Database, weight.Webserver, weight.Threshold)
	return err
}

func (p *postgresHandler) GetThreshold() (float32, error) {
	row := p.db.QueryRow("SELECT threshold FROM weight LIMIT 1")
	var threshold float32
	err := row.Scan(&threshold)
	if err != nil {
		return threshold, err
	}
	return threshold, nil
}

func (p *postgresHandler) SetThreshold(t float32) error {
	_, err := p.db.Exec("INSERT INTO weight (threshold) VALUES ($1) ON CONFLICT DO UPDATE", t)
	return err
}

func (p *postgresHandler) GetVMInfo(id string) (*data.VMInstance, error) {
	row := p.db.QueryRow("SELECT id, name, os, language, database, webserver FROM vminfo WHERE id = $1", id)

	var vm data.VMInstance
	var languagesStr, databasesStr, webserversStr string

	err := row.Scan(&vm.ID, &vm.Name, &vm.OS, &languagesStr, &databasesStr, &webserversStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no VM instance found with ID: %s", id)
		}
		return nil, fmt.Errorf("error scanning database row: %v", err)
	}

	err = json.Unmarshal([]byte(languagesStr), &vm.Software.Languages)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling languages data: %v", err)
	}

	err = json.Unmarshal([]byte(databasesStr), &vm.Software.Databases)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling databases data: %v", err)
	}

	err = json.Unmarshal([]byte(webserversStr), &vm.Software.Webservers)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling webservers data: %v", err)
	}

	return &vm, nil
}

func (p *postgresHandler) SetVMInfo(v data.VMInstance) error {
	languageText, err := json.Marshal(v.Software.Languages)
	if err != nil {
		return err
	}

	databaseText, err := json.Marshal(v.Software.Databases)
	if err != nil {
		return err
	}

	webserverText, err := json.Marshal(v.Software.Webservers)
	if err != nil {
		return err
	}

	statement, err := p.db.Prepare("INSERT INTO vminfo (id, language, database, webserver) VALUES ($1, $2, $3, $4, $5")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(v.ID, languageText, databaseText, webserverText)
	if err != nil {
		return err
	}

	return nil
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
