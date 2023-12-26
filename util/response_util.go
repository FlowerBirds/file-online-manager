package util

import (
	"encoding/json"
	"file-online-manager/model"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Error(w http.ResponseWriter, err error) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		i := strings.LastIndex(file, "/")
		// 将调用栈信息与日志消息一起输出
		log.Printf("[%s line:%d]: %s", file[i+1:], line, err)
	} else {
		log.Println(err)
	}
	response := model.Response{Code: 500, Message: err.Error(), Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonResponse)
}

func Println(v ...any) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		i := strings.LastIndex(file, "/")
		// 将调用栈信息与日志消息一起输出
		log.Printf("[%s line:%d]: %s", file[i+1:], line, fmt.Sprintln(v...))
	} else {
		log.Println(v)
	}
}
