package controller

import (
	// "fmt"
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/model"
	"github.com/tangyang/simple-http-server/service"
	"net/http"
)

var userService *service.UserService = &service.UserService{}

func addUser(c *config.Config, w http.ResponseWriter, r *http.Request) interface{} {
	m := parseParameter(r)
	name := m["name"].(string)
	if len(name) <= 0 {
		return model.Result{Code: http.StatusBadRequest, Message: "Name parameter is required! "}
	}
	b, err := userService.AddUser(c, &model.User{Name: name})
	if !b {
		return model.Result{Code: http.StatusBadRequest, Message: "Name already exists! "}
	}
	if err != nil {
		return model.Result{Code: http.StatusInternalServerError, Message: "Something is wrong with server. "}
	}
	return model.Result{Code: http.StatusOK, Message: "", Data: userService.GetUserByName(c, name)}
}

func getAllUsers(c *config.Config, w http.ResponseWriter, r *http.Request) interface{} {
	users := userService.GetAllUsers(c)
	return model.Result{Code: http.StatusOK, Message: "", Data: users}
}
