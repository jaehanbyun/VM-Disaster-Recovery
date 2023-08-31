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
