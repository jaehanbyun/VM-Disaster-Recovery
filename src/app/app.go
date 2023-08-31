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

	return a
}
