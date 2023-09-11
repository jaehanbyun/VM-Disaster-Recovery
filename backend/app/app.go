package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jaehanbyun/VM-Disaster-Recovery/common"
	"github.com/jaehanbyun/VM-Disaster-Recovery/data"
	"github.com/jaehanbyun/VM-Disaster-Recovery/model"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

type Apphandler struct {
	http.Handler
	db model.DBHandler
}

var (
	rd *render.Render
	// PortForwarded Openstack VM IP
	Similaritys map[string]int
)

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://*")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token ,Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func (a *Apphandler) getInstances(w http.ResponseWriter, r *http.Request) {
	token := GetToken()
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/compute/v2.1/compute/servers/detail", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var openStackResp data.OpenStackResponse
	json.Unmarshal(body, &openStackResp)

	var finalServers []data.VMInstance
	for _, server := range openStackResp.Servers {
		volumeIDs := server.OsExtendedVolumesVolumesAttached
		var os string
		languages, databases, webservers := []string{}, []string{}, []string{}
		for _, volumeID := range volumeIDs {
			metadata, err := common.GetVolumeMetadata(token, volumeID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			switch metadata.Type {
			case "os":
				os = metadata.Content
			case "database":
				databases = append(databases, metadata.Content)
			case "webserver":
				webservers = append(webservers, metadata.Content)
			case "language":
				languages = append(languages, metadata.Content)
			}
		}

		finalServer := data.VMInstance{
			ID: server.ID,
			OS: os,
			Software: data.Software{
				Languages:  languages,
				Databases:  databases,
				Webservers: webservers,
			},
		}
		finalServers = append(finalServers, finalServer)
	}

	finalJson, err := json.Marshal(map[string][]data.VMInstance{"servers": finalServers})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	rd.JSON(w, http.StatusOK, finalJson)
}

func (a *Apphandler) getInstanceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	instance, err := a.db.GetVMInfo(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching VM Info: %s", err), http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(instance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling response: %s", err), http.StatusInternalServerError)
		return
	}

	rd.JSON(w, http.StatusOK, respBytes)
}

func (a *Apphandler) getVolumes(w http.ResponseWriter, r *http.Request) {
	token := GetToken()
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/v3"+common.ProjectId+"/volumes/detail", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var volumeList data.VolumeListResponse
	json.Unmarshal(body, &volumeList)

	summaryMap := make(map[string][]data.SummaryDetail)

	for _, v := range volumeList.Volumes {
		vType := v.Metadata.Type
		summaryDetail := data.SummaryDetail{Name: v.Name, ID: v.ID, Content: v.Metadata.Content}
		summaryMap[vType] = append(summaryMap[vType], summaryDetail)
	}

	finalOutput := map[string]map[string][]data.SummaryDetail{
		"volumes": summaryMap,
	}

	jsonData, err := json.Marshal(finalOutput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rd.JSON(w, http.StatusOK, jsonData)
}

func GetToken() string {
	payload := data.Payload{
		Auth: data.Auth{
			Identity: data.Identity{
				Methods: []string{"password"},
				Password: data.Password{
					User: data.User{
						Name: "admin",
						Domain: data.Domain{
							Name: "Default",
						},
						Password: "0000",
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	buff := bytes.NewBuffer(body)
	resp, err := http.Post(common.BaseOpenstackUrl+"/identity/v3/auth/tokens", "application/json", buff)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		token := resp.Header.Get("X-Subject-Token")
		return token
	}

	return ""
}

func GetVolumeMetadata(token string, id string) (data.Metadata, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/volume/v3/"+common.ProjectId+"/volumes/"+id, nil)
	if err != nil {
		return data.Metadata{}, fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Add("X-Auth-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return data.Metadata{}, fmt.Errorf("error making request: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data.Metadata{}, fmt.Errorf("received non 200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data.Metadata{}, fmt.Errorf("error reading response body: %s", err)
	}

	var volume data.VolumeResponse
	if err := json.Unmarshal(body, &volume); err != nil {
		return data.Metadata{}, fmt.Errorf("error unmarshaling JSON: %s", err)
	}

	return volume.Volume.Metadata, nil
}

func MakeHandler() *Apphandler {
	rd = render.New()
	r := mux.NewRouter()
	r.Use(enableCORS)

	neg := negroni.Classic()
	neg.UseHandler(r)

	a := &Apphandler{
		Handler: neg,
		db:      model.NewDBHandler(),
	}

	r.HandleFunc("/volumes", a.getVolumes).Methods("GET")
	r.HandleFunc("/instance", a.getInstances).Methods("GET")
	r.HandleFunc("/instance/{id}", a.getInstanceByID).Methods("GET")

	return a
}
