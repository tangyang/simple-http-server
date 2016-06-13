package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func parseParameter(r *http.Request) map[string]interface{} {
	result, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	fmt.Printf("%s\n", result)
	var f interface{}
	json.Unmarshal(result, &f)
	return f.(map[string]interface{})
}
