package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jaehanbyun/VM-Disaster-Recovery/common"
	"github.com/jaehanbyun/VM-Disaster-Recovery/data"
	_ "github.com/lib/pq"
)

type postgresHandler struct {
	db *sql.DB
}

func (p *postgresHandler) Close() {
	p.db.Close()
}

func (p *postgresHandler) Init() error {
	token := common.GetToken()
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/image/v2/images", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching image info: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var imageResp data.ImageListResponse
	err = json.Unmarshal(body, &imageResp)
	if err != nil {
		return fmt.Errorf("error unmarshalling response: %v", err)
	}

	statement, err := p.db.Prepare("INSERT INTO osinfo (id, name) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}

	for _, image := range imageResp.Images {
		_, err := statement.Exec(image.ID, image.Name)
		if err != nil {
			return fmt.Errorf("error inserting record: %v", err)
		}
	}

	err = p.SetVMsInfo()
	if err != nil {
		return fmt.Errorf("error setting vms info: %v", err)
	}

	weight := &data.Weight{
		Language:  1,
		Database:  1,
		Webserver: 1,
		Threshold: 0.8,
	}
	err = p.SetWeight(*weight)
	if err != nil {
		return fmt.Errorf("error setting weight: %v", err)
	}

	return nil
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

func (p *postgresHandler) GetVMsInfo() ([]*data.VMInstance, error) {
	rows, err := p.db.Query("SELECT id, name, flavorid, os, language, database, webserver FROM vminfo")
	if err != nil {
		return nil, fmt.Errorf("error querying vminfo: %v", err)
	}
	defer rows.Close()

	var vmInstances []*data.VMInstance

	for rows.Next() {
		var vm data.VMInstance
		var languagesStr, databasesStr, webserversStr string

		if err := rows.Scan(&vm.ID, &vm.Name, &vm.FlavorID, &vm.OS, &languagesStr, &databasesStr, &webserversStr); err != nil {
			return nil, fmt.Errorf("error scanning databases: %v", err)
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

		vmInstances = append(vmInstances, &vm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %v", err)
	}

	return vmInstances, nil
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

func (p *postgresHandler) SetVMsInfo() error {
	token := common.GetToken()
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/compute/v2.1/servers/detail", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching instance info: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var openStackResp data.OpenStackResponse
	json.Unmarshal(body, &openStackResp)

	var vms []data.VMInstance
	for _, server := range openStackResp.Servers {
		serverName := server.Name
		flavorID := server.Flavor.ID
		volumeIDs := server.OsExtendedVolumesVolumesAttached
		os, err := p.GetImageName(server.OS.ID)
		if err != nil {
			return fmt.Errorf("error getting os name: %s", err)
		}

		var languages, databases, webservers []data.Volume
		for _, volumeID := range volumeIDs {
			metadata, err := common.GetVolumeMetadata(token, volumeID.ID)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("error fetching volume metadata: %s", err))
			}
			vol := data.Volume{
				ID:      volumeID.ID,
				Content: metadata.Content,
			}
			switch metadata.Type {
			case "database":
				databases = append(databases, vol)
			case "webserver":
				webservers = append(webservers, vol)
			case "language":
				languages = append(languages, vol)
			}
		}

		vm := data.VMInstance{
			ID:       server.ID,
			FlavorID: flavorID,
			Name:     serverName,
			OS:       os,
			Software: data.Software{
				Languages:  languages,
				Databases:  databases,
				Webservers: webservers,
			},
		}
		vms = append(vms, vm)
	}

	statement, err := p.db.Prepare("INSERT INTO vminfo (id, name, flavorid, os, language, database, webserver) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}

	for _, vm := range vms {
		languagesJSON, _ := json.Marshal(vm.Software.Languages)
		databasesJSON, _ := json.Marshal(vm.Software.Databases)
		webserversJSON, _ := json.Marshal(vm.Software.Webservers)

		_, err := statement.Exec(vm.ID, vm.Name, vm.FlavorID, vm.OS, languagesJSON, databasesJSON, webserversJSON)
		if err != nil {
			return fmt.Errorf("error inserting VM record: %v", err)
		}
	}

	return nil
}

func (p *postgresHandler) GetImageName(id string) (string, error) {
	row := p.db.QueryRow("SELECT name FROM osinfo WHERE id = $1", id)
	var osName string
	err := row.Scan(&osName)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("error scanning osinfo: %v", err)
	}
	return osName, nil
}

func newPostgresHandler() DBHandler {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "vms",
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
				id INT PRIMARY KEY,
				language NUMERIC,
				database NUMERIC,
				webserver NUMERIC,
				threshold NUMERIC
			);`)
	_, err = createWeight.Exec()
	if err != nil {
		panic(err)
	}

	createVMInfo, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS vminfo (
			id TEXT PRIMARY KEY,
			name TEXT,
			flavorid TEXT,
			os TEXT,
			language JSON,
			database JSON,
			webserver JSON
		);`)
	_, err = createVMInfo.Exec()
	if err != nil {
		panic(err)
	}

	createOSInfo, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS osinfo (
			id TEXT PRIMARY KEY,
			name TEXT
		);`)
	_, err = createOSInfo.Exec()
	if err != nil {
		panic(err)
	}

	return &postgresHandler{database}
}
