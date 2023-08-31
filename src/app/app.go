package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
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
	baseOpenstackUrl = "10.125.70.26:8888"
	projectId        = "example"
	Similaritys      map[string]int
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

func (a *Apphandler) getVolumes(w http.ResponseWriter, r *http.Request) {
	token := getToken()
	req, err := http.NewRequest("GET", "http://"+baseOpenstackUrl+"/v3"+projectId+"/volumes/detail", nil)
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

	summaryMap := make(map[string]data.Summary)

	for _, v := range volumeList.Volumes {
		vType := v.Metadata.Type
		if _, ok := summaryMap[vType]; !ok {
			summaryMap[vType] = data.Summary{}
		}
		summaryMap[vType] = data.Summary{ID: append(summaryMap[vType].ID, v.ID)}
	}

	finalOutput := map[string]map[string]data.Summary{
		"volumes": summaryMap,
	}

	jsonData, err := json.Marshal(finalOutput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rd.JSON(w, http.StatusOK, jsonData)
}

func getToken() string {
	payload := data.Payload{
		Auth: data.Auth{
			Identity: data.Identity{
				Methods: []string{"password"},
				Password: data.Password{
					User: data.User{
						Name:     "admin",
						Domain:   "Default",
						Password: "0000",
					},
				},
			},
			Scope: data.Scope{
				System: data.System{
					All: true,
				},
			},
		},
	}

	body, err := json.Marshal(payload)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	buff := bytes.NewBuffer(body)
	resp, err := http.Post("http://"+baseOpenstackUrl+"identity/v3/auth/tokens", "application/json", buff)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		token := resp.Header.Get("X-Subject-Token")
		return token
	}

	return ""
}

func findSimilarVM(id string) *data.VMInstance {
	return &data.VMInstance{}
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

	return a
}
