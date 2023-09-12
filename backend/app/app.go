package app

import (
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

type AppHandler struct {
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

func (a *AppHandler) getInstances(w http.ResponseWriter, r *http.Request) {
	vms, err := a.db.GetVMsInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting vm info: %v", err), http.StatusInternalServerError)
		return
	}

	rd.JSON(w, http.StatusOK, vms)
}

func (a *AppHandler) getInstanceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	vm, err := a.db.GetVMInfo(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rd.JSON(w, http.StatusOK, vm)
}

func (a *AppHandler) getVolumes(w http.ResponseWriter, r *http.Request) {
	token := common.GetToken()
	req, err := http.NewRequest("GET", common.BaseOpenstackUrl+"/volume/v3/"+common.ProjectId+"/volumes/detail", nil)
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

	rd.JSON(w, http.StatusOK, finalOutput)
}

func MakeHandler() *AppHandler {
	rd = render.New()
	r := mux.NewRouter()
	r.Use(enableCORS)

	neg := negroni.Classic()
	neg.UseHandler(r)

	a := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(),
	}

	r.HandleFunc("/volumes", a.getVolumes).Methods("GET")
	r.HandleFunc("/instance", a.getInstances).Methods("GET")
	r.HandleFunc("/instance/{id}", a.getInstanceByID).Methods("GET")

	return a
}
