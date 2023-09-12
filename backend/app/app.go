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

func (a *AppHandler) createInstance(w http.ResponseWriter, r *http.Request) {
	var instanceReq data.InstanceRequest

	err := json.NewDecoder(r.Body).Decode(&instanceReq)
	if err != nil {
		http.Error(w, "Failed to parse requset body", http.StatusBadRequest)
		return
	}

	fmt.Println(instanceReq.Volumes)

	err = generateTerraformScript(instanceReq.Name, instanceReq.OS, instanceReq.Ram, instanceReq.Vcpus, instanceReq.Disk, instanceReq.Volumes)
	if err != nil {
		http.Error(w, "Failed to generate terraform script", http.StatusInternalServerError)
		return
	}

	err = runTerraformApply()
	if err != nil {
		http.Error(w, "Failed to run terraform script", http.StatusInternalServerError)
		return
	}

	rd.Text(w, http.StatusOK, "OK")
}

func (a *AppHandler) recoverInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	targetVM, err := a.db.GetVMInfo(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching VM with ID: %s. Error: %v", id, err), http.StatusInternalServerError)
		return
	}

	allVMs, err := a.db.GetVMsInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching all VMs: %v", err), http.StatusInternalServerError)
		return
	}

	weights, err := a.db.GetWeight()
	maxSimilarity := 0.0
	var mostSimilarVM *data.VMInstance

	for _, vm := range allVMs {
		if vm.ID == targetVM.ID {
			continue
		}
		similarity := calculateSimilarity(weights, *targetVM, *vm)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
			mostSimilarVM = vm
		}
	}

	if maxSimilarity >= float64(weights.Threshold) {
		var nonOverlappingVolumes []string
		for _, volume := range targetVM.Software.Languages {
			if !contains(mostSimilarVM.Software.Languages, volume) {
				nonOverlappingVolumes = append(nonOverlappingVolumes, volume.ID)
			}
		}
		for _, volume := range targetVM.Software.Databases {
			if !contains(mostSimilarVM.Software.Databases, volume) {
				nonOverlappingVolumes = append(nonOverlappingVolumes, volume.ID)
			}
		}
		for _, volume := range targetVM.Software.Webservers {
			if !contains(mostSimilarVM.Software.Webservers, volume) {
				nonOverlappingVolumes = append(nonOverlappingVolumes, volume.ID)
			}
		}

		token := common.GetToken()

		for _, volumeID := range nonOverlappingVolumes {
			body, err := json.Marshal(makeVolumeAttachmentReq(volumeID))
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to marshal volume attachment data: %v", err), http.StatusInternalServerError)
				return
			}
			req, err := http.NewRequest("POST", common.BaseOpenstackUrl+"/compute/v2.1/servers/"+mostSimilarVM.ID+"/os-volume_attachments", bytes.NewBuffer(body))
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating new request: %s", err), http.StatusInternalServerError)
				return
			}
			req.Header.Set("X-Auth-Token", token)
			req.Header.Set("content-type", "application/json")
			// Openstack Nova Compute API Version 2.60 Include
			req.Header.Set("Openstack-API-Version", "compute 2.60")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, fmt.Sprintf("error sending volume attachment request: %s", err), http.StatusInternalServerError)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				http.Error(w, "failed to attach volume", http.StatusInternalServerError)
				return
			}
		}
	} else {
		newVMName := targetVM.Name + "-new"
		var volumes []string

		for _, langVolume := range targetVM.Software.Languages {
			volumes = append(volumes, langVolume.ID)
		}
		for _, dbVolume := range targetVM.Software.Databases {
			volumes = append(volumes, dbVolume.ID)
		}
		for _, wsVolume := range targetVM.Software.Webservers {
			volumes = append(volumes, wsVolume.ID)
		}

		err := recreateTerraformScript(newVMName, targetVM.OS, targetVM.FlavorID, volumes)
		if err != nil {
			rd.Text(w, http.StatusInternalServerError, fmt.Sprintf("Failed to recreate VM: %s", err))
			return
		}
	}

	rd.Text(w, http.StatusOK, fmt.Sprintf("Created a new VM with base of VM with ID %s", targetVM.ID))
}

func makeVolumeAttachmentReq(volumeId string) data.VolumeAttachmentsRequest {
	volumeAttachmentReq := data.VolumeAttachmentsRequest{
		VolumeAttachment: data.VolumeAttachment{
			VolumeID: volumeId,
		},
	}

	return volumeAttachmentReq
}

func contains(volumes []data.Volume, vol data.Volume) bool {
	for _, v := range volumes {
		if v.ID == vol.ID && v.Content == vol.Content {
			return true
		}
	}
	return false
}

func calculateSimilarity(weight data.Weight, source, target data.VMInstance) float64 {
	if source.OS != target.OS {
		return 0.0
	}

	totalWeight := 0.0
	similarWeight := 0.0

	for _, lang := range source.Software.Languages {
		totalWeight += float64(weight.Language)
		if contains(target.Software.Languages, lang) {
			similarWeight += float64(weight.Language)
		}
	}

	for _, db := range source.Software.Databases {
		totalWeight += float64(weight.Database)
		if contains(target.Software.Databases, db) {
			similarWeight += float64(weight.Database)
		}
	}

	for _, ws := range source.Software.Webservers {
		totalWeight += float64(weight.Webserver)
		if contains(target.Software.Webservers, ws) {
			similarWeight += float64(weight.Webserver)
		}
	}

	return similarWeight / totalWeight
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
	r.HandleFunc("/instance", a.createInstance).Methods("POST")
	r.HandleFunc("/instance/{id}", a.getInstanceByID).Methods("GET")
	r.HandleFunc("/instance/{id}/recover", a.recoverInstance).Methods("POST")

	err := a.db.Init()

	if err != nil {
		panic(err)
	}

	return a
}
