package app

import (
	"net/http"

	"github.com/gorilla/mux"
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

func MakeHandler() *Apphandler {
	rd = render.New()
	r := mux.NewRouter()

	neg := negroni.Classic()
	neg.UseHandler(r)

	a := &Apphandler{
		Handler: neg,
		db:      model.NewDBHandler(),
	}

	return a
}
