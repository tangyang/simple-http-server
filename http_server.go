package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/controller"
	"net/http"
)

type dispatcher struct {
	handler http.Handler
}

func (d *dispatcher) SetHandler(handler http.Handler) {
	d.handler = handler
}

func initHttpServer(conf *config.Config) {

	r := mux.NewRouter()
	controller.InitRouters(r, conf)
	go func() {
		err := http.ListenAndServe(":8000", r)
		fmt.Println(err)
	}()
}
