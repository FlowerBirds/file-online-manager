package util

import (
	"encoding/json"
	"file-online-manager/model"
	"log"
	"net/http"
)

func Error(w http.ResponseWriter, err error) {
	log.Println(err)
	response := model.Response{Code: 400, Message: err.Error(), Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonResponse)
}
