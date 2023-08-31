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
	Scope    Scope    `json:"scope"`
}

type Scope struct {
	System System `json:"system"`
}

type System struct {
	All bool `json:"all"`
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
	Domain   string `json:"domain"`
	Password string `json:"password"`
}

type VolumeDetail struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type VolumeListResponse struct {
	Volumes []VolumeDetail `json:"volumes"`
}

type Summary struct {
	ID []string `json:"id"`
}
