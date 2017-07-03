package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Ajax struct {
	Success bool
	Message string
}

func ajaxResponse(w http.ResponseWriter, r *http.Request, success bool, data interface{}, err string) {

	marshaled, _ := json.Marshal(data)

	payload := string(marshaled)

	if len(payload) == 0 {
		payload = "\"\""
	}

	fmt.Fprintf(w, "{\"success\": %v, \"data\": %v, \"error\": \"%v\"}", success, payload, err)
}
