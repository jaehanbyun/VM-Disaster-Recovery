package data

type Weight struct {
	Language  float32 `json:"language"`
	Database  float32 `json:"database"`
	Webserver float32 `json:"webserver"`
	Threshold float32 `json:"thtreshold"`
}

type Software struct {
	Languages  []Volume `json:"languages"`
	Databases  []Volume `json:"databases"`
	Webservers []Volume `json:"webservers"`
}

type Volume struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type VMInstance struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	FlavorID string   `json:"flavor_id"`
	OS       string   `json:"os"`
	Software Software `json:"software"`
}

type Payload struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Identity Identity `json:"identity"`
}

type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

type Password struct {
	User User `json:"user"`
}

type User struct {
	Name     string `json:"name"`
	Domain   Domain `json:"domain"`
	Password string `json:"password"`
}

type Domain struct {
	Name string `json:"name"`
}

type VolumeDetail struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type VolumeResponse struct {
	Volume VolumeDetail `json:"volume"`
}

type VolumeListResponse struct {
	Volumes []VolumeDetail `json:"volumes"`
}

type SummaryDetail struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ServerDetail struct {
	ID                               string                 `json:"id"`
	Name                             string                 `json:"name"`
	Flavor                           Flavor                 `json:"flavor"`
	OS                               ImageDetail            `json:"image"`
	OsExtendedVolumesVolumesAttached []AttachVolumeID       `json:"os-extended-volumes:volumes_attached"`
	Metadata                         map[string]interface{} `json:"metadata"`
}

type Flavor struct {
	ID string `json:"id"`
}

type AttachVolumeID struct {
	ID string `json:"id"`
}

type OpenStackResponse struct {
	Servers []ServerDetail `json:"servers"`
}

type InstanceRequest struct {
	Name    string   `json:"name"`
	Ram     int      `json:"ram"`
	Vcpus   int      `json:"vcpus"`
	Disk    int      `json:"disk"`
	OS      string   `json:"os_name"`
	Volumes []string `json:"volumes"`
}

type ImageListResponse struct {
	Images []ImageDetail `json:"images"`
}
type ImageDetail struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type VolumeAttachmentsRequest struct {
	VolumeAttachment VolumeAttachment `json:"volumeAttachment"`
}

type VolumeAttachment struct {
	VolumeID string `json:"volumeId"`
}
