package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jaehanbyun/VM-Disaster-Recovery/data"
)

var (
	// PortForwarded Openstack VM IP
	BaseOpenstackUrl = "http://10.125.70.26:8889"
	ProjectId        = "66d5c0c9a8464550906e95d0b23c161f"
)

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
	resp, err := http.Post(BaseOpenstackUrl+"/identity/v3/auth/tokens", "application/json", buff)

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
	req, err := http.NewRequest("GET", BaseOpenstackUrl+"/volume/v3/"+ProjectId+"/volumes/"+id, nil)
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
