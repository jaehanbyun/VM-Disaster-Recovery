package data

type Weight struct {
	Language  float32 `json:"language"`
	Database  float32 `json:"database"`
	Webserver float32 `json:"webserver"`
	Threshold float32 `json:"thtreshold"`
}

type Software struct {
	Languages  []string `json:"languages"`
	Databases  []string `json:"databases"`
	Webservers []string `json:"webservers"`
}

type VMInstance struct {
	ID       string   `json:"id"`
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
	OsExtendedVolumesVolumesAttached []string               `json:"os_extended_volumes_attached:volumes_attached"`
	Metadata                         map[string]interface{} `json:"metadata"`
}

type OpenStackResponse struct {
	Servers []ServerDetail `json:"servers"`
}

