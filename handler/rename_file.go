package handler

import (
	"encoding/json"
	"file-online-manager/model"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func RenameFileHandler(w http.ResponseWriter, r *http.Request) {

	// 请求类型为application/json中获取参数，而不是form表单
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request body", http.StatusBadRequest)
		return
	}
	var requestData model.RequestFileData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	filePath := requestData.Path
	newName := requestData.Name

	if filePath == "" || newName == "" {
		response := model.Response{Code: 400, Message: "Missing path or new_name parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	dir := filepath.Dir(filePath)
	newFilePath := dir + "/" + newName
	err = os.Rename(filePath, newFilePath)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to rename file", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "File renamed successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
