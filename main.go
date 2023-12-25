package main

import (
	"encoding/json"
	"file-online-manager/handler"
	"file-online-manager/handler/k8sservice"
	"file-online-manager/model"
	"file-online-manager/util"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var root = "."
var loginUsername = ""
var loginPassword = ""

func main() {
	router := mux.NewRouter()
	contextPath := "/fm"
	path := os.Getenv("CONTEXT_PATH")
	if len(path) > 0 {
		contextPath = path
	}
	rootPath := os.Getenv("ROOT_PATH")
	if len(rootPath) > 0 {
		root = rootPath
	}
	_, err := os.Stat(root)
	if err != nil {
		err = os.Mkdir(root, os.ModePerm)
		if err != nil {
			fmt.Printf("创建目录异常 -> %v\n", err)
			return
		}
	}
	log.Println("server manage root path: " + root)
	log.Println("server use context path: " + contextPath)

	initAuth()
	// 增加拦截器
	router.Use(accessLogMiddleware, authenticationMiddleware)
	handler.RootPath = root

	router.HandleFunc(contextPath+"api/manager/file/delete", handler.DeleteFileHandler).Methods("DELETE")
	router.HandleFunc(contextPath+"api/manager/file/rename", handler.RenameFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/file/list", listFileHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/copy", handler.CopyFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/file/upload", handler.UploadLagerFileHandler).Methods("POST", "GET")
	router.HandleFunc(contextPath+"api/manager/file/unzip", handler.UnzipFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/file/download", handler.DownloadHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/zip/view", handler.ViewZipFileHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/zip/release", handler.ReleaseZipFileHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/content", handler.TextFileViewHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/content", handler.TextFileSaveHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/folder/list", listFolderHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/folder/delete", deleteFolderHandler).Methods("DELETE")
	router.HandleFunc(contextPath+"api/manager/folder/rename", renameFolderHandler).Methods("PUT")
	router.HandleFunc(contextPath+"api/manager/folder/copy", copyFolderHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/folder/create", createFolderHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/folder/zip", zipFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/k8s/list-pods", k8sservice.ListPodHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/k8s/restart-pod", k8sservice.RestartPodHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/k8s/list-namespace", k8sservice.ListNamespaceHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/k8s/pod-stream-logs", k8sservice.PodStreamLogHandler).Methods("GET")

	router.PathPrefix(contextPath + "").Handler(http.StripPrefix(contextPath, http.FileServer(http.Dir("./static/"))))
	log.Println("server started at port 8080")
	http.ListenAndServe(":8080", router)
}

/**
 * 根据设置的策略，进行认证初始化
 */
func initAuth() {
	manageUsername := os.Getenv("MANAGE_USERNAME")
	managePassword := os.Getenv("MANAGE_PASSWORD")
	manageSecurity := os.Getenv("MANAGE_SECURITY")
	if manageUsername == "" || manageSecurity == "true" || manageSecurity == "" {
		loginUsername = util.GenToken(32)
		log.Println("use security user: " + loginUsername)
	} else {
		loginUsername = manageUsername
	}
	if managePassword == "" || manageSecurity == "true" || manageSecurity == "" {
		loginPassword = util.GenToken(128)
		log.Println("use security token: " + loginPassword)
	} else {
		loginPassword = managePassword
	}
	if manageSecurity == "true" || manageSecurity == "" {
		expireTimeStr := os.Getenv("EXPIRE_TIME")
		expireTime, err := strconv.Atoi(expireTimeStr)
		if err != nil || expireTime < 1 {
			expireTime = 24
		}
		log.Printf("use security mode, user token will be update %d hours. \n", expireTime)
		ticker := time.NewTicker(time.Duration(expireTime) * time.Hour)
		go func() {
			for {
				select {
				case <-ticker.C:
					// 更新token
					loginUsername = util.GenToken(32)
					log.Println("use security user: " + loginUsername)
					loginPassword = util.GenToken(128)
					log.Println("use security token: " + loginPassword)
				}
			}
		}()
	}
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 在每个请求之前进行身份验证
		username, password, ok := r.BasicAuth()
		if !ok || !checkAuth(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// 继续执行下一个处理函数
		next.ServeHTTP(w, r)
	})
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Path
		log.Println("access uri:", uri)

		// 继续执行下一个处理函数
		next.ServeHTTP(w, r)
	})
}

func listFileHandler(w http.ResponseWriter, r *http.Request) {
	handler.ListFileHandler(root, w, r)
}

func listFolderHandler(w http.ResponseWriter, r *http.Request) {

	folders := []model.File{}
	path := r.FormValue("path")
	if len(path) == 0 {
		path = root
	}
	// 解决文件夹展示位置为当前进程启动位置而不是环境变量设置的位置
	if strings.Index(path, ".") == 0 {
		path = root + "/" + path
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to list folders", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, model.File{Name: file.Name(), Path: path + "/" + file.Name(), IsDir: true, Id: uuid.New().String()})
		}
	}
	response := model.Response{Code: 200, Message: "Folders listed successfully", Data: folders}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func deleteFolderHandler(w http.ResponseWriter, r *http.Request) {

	folderPath := r.FormValue("path")
	if folderPath == "" {
		response := model.Response{Code: 400, Message: "Missing path parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	err := os.RemoveAll(folderPath)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to delete folder", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "Folder deleted successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func renameFolderHandler(w http.ResponseWriter, r *http.Request) {

	folderPath := r.FormValue("path")
	newName := r.FormValue("new_name")
	if folderPath == "" || newName == "" {
		response := model.Response{Code: 400, Message: "Missing path or new_name parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	err := os.Rename(folderPath, newName)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to rename folder", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "Folder renamed successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func copyFolderHandler(w http.ResponseWriter, r *http.Request) {

	filePath := r.FormValue("path")
	copyName := r.FormValue("name")
	if filePath == "" || copyName == "" {
		response := model.Response{Code: 400, Message: "Missing path or name parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	fileInfo, err := os.Stat(filePath)
	// check if folder exists
	if os.IsNotExist(err) {
		response := model.Response{Code: 500, Message: "Failed to check folder", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	if fileInfo.IsDir() {
		response := model.Response{Code: 500, Message: "Not support to copy folder", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}

	dir := filepath.Dir(filePath)
	newPath := dir + "/" + copyName
	if _, err := os.Stat(newPath); err == nil {
		response := model.Response{Code: 500, Message: "The target file exists", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}

	copyFile(filePath, newPath)

	response := model.Response{Code: 200, Message: "File copied successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// 从环境变量中获取用户名和密码
func checkAuth(username string, password string) bool {
	if username == "" || password == "" {
		return false
	}
	return username == loginUsername && password == loginPassword
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Sync()
	if err != nil {
		return err
	}
	s, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, s.Mode())
	if err != nil {
		return err
	}
	return nil
}

func createFolderHandler(w http.ResponseWriter, r *http.Request) {

	folderPath := r.FormValue("path")
	if folderPath == "" || folderPath == "." || folderPath == "/" {
		response := model.Response{Code: 400, Message: "Missing path parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	if strings.Index(folderPath, "./") == 0 {
		folderPath = root + "/" + folderPath
	}
	// check if folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// create new folder
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			response := model.Response{Code: 500, Message: "Failed to create new folder", Data: nil}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonResponse)
			return
		}
	} else {
		response := model.Response{Code: 400, Message: "Folder already exists", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "Folder created successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func zipFileHandler(w http.ResponseWriter, r *http.Request) {
	// 请求类型为application/json中获取参数，而不是form表单
	body, err := io.ReadAll(r.Body)
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

	// filePath := r.FormValue("path")
	// copyName := r.FormValue("name")
	filePath := requestData.Path
	//fileName := requestData.Name

	if filePath == "" {
		response := model.Response{Code: 400, Message: "Missing path parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	if strings.Index(filePath, "./") == 0 {
		filePath = root + "/" + filePath
	}
	// 获取文件所在的目录，并解压到当前目录
	fileName := filepath.Base(filePath)
	fileDir := filepath.Dir(filePath)
	var cmdErr error = nil
	// 切换当前工作目录，执行完需要切换回去
	currentPath, _ := os.Getwd()
	os.Chdir(fileDir)
	cmdErr = util.ExecuteCommand("zip", "-r", fileName+".zip", fileName)
	os.Chdir(currentPath)
	if cmdErr != nil {
		response := model.Response{Code: 400, Message: cmdErr.Error(), Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		log.Println("zip failed", cmdErr)
		return
	}

	response := model.Response{Code: 200, Message: "File zip successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
