package main

import (
	"net/http"

	"github.com/jaehanbyun/VM-Disaster-Recovery/app"
)

func main() {
	r := app.MakeHandler()

	http.ListenAndServe(":8000", r)
}
