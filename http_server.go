package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/controller"
	"net/http"
	"strings"
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
		port := strings.Join([]string{"0.0.0.0", conf.HttpPort}, ":")
		err := http.ListenAndServe(port, r)
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("Http server is initialized... ")
}
