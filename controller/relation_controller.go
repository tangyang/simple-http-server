package controller

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/model"
	"github.com/tangyang/simple-http-server/service"
	"github.com/tangyang/simple-http-server/to"
	"net/http"
	"strconv"
	"strings"
)

var relationService *service.RelationService = &service.RelationService{}

func getAllRelations(c *config.Config, w http.ResponseWriter, r *http.Request) interface{} {
	vars := mux.Vars(r)
	userId, _ := strconv.ParseInt(vars["userId"], 10, 64)
	relations := relationService.GetRelations(c, userId)
	return model.Result{Code: http.StatusOK, Message: "", Data: to.NewRelationToArray(relations)}
}

func addNewRelation(c *config.Config, w http.ResponseWriter, r *http.Request) interface{} {
	m := parseParameter(r)
	status, err := parseStatus(m["state"].(string))
	if err != nil {
		return model.Result{Code: http.StatusBadRequest, Message: "Bad parameter status"}
	}

	vars := mux.Vars(r)
	userId, _ := strconv.ParseInt(vars["userId"], 10, 64)
	otherUserId, _ := strconv.ParseInt(vars["otherUserId"], 10, 64)

	relation := &model.Relation{Userid: userId, Otheruserid: otherUserId, Status: status}

	relationService.AddRelation(c, relation)

	return model.Result{Code: http.StatusOK, Message: "", Data: to.NewRelationTo(relation)}
}

func parseStatus(status string) (model.RelationStatus, error) {
	if strings.EqualFold(status, "liked") {
		return model.RelationLike, nil
	} else if strings.EqualFold(status, "disliked") {
		return model.RelationDislike, nil
	} else {
		return -1, errors.New(fmt.Sprintf("unrecognized status parameter, %s", status))
	}
}
