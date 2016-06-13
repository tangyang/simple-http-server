package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/tangyang/simple-http-server/config"
	"net/http"
)

type handler func(c *config.Config, w http.ResponseWriter, r *http.Request) interface{}

var routes = map[string]map[string]handler{
	"GET": {
		"/users": getAllUsers,
	},
	"POST": {
		"/users": addUser,
	},
	"PUT":    {},
	"DELETE": {},
}

func InitRouters(r *mux.Router, c *config.Config) {
	for method, mappings := range routes {
		for route, fct := range mappings {

			localRoute := route
			localFct := fct

			wrap := func(w http.ResponseWriter, r *http.Request) {
				result := localFct(c, w, r)
				w.Header().Set("Content-type", "application/json")
				json.NewEncoder(w).Encode(result)
			}
			localMethod := method

			r.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}
}
