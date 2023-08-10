package main

import (
	"encoding/json"
	"file-online-manager/handler"
	"file-online-manager/model"
	"file-online-manager/util"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var root = "."
var loginUsername = ""
var loginPassword = ""

func main() {
	router := mux.NewRouter()
	contextPath := "/"
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

	router.HandleFunc(contextPath+"api/manager/file/delete", deleteFileHandler).Methods("DELETE")
	router.HandleFunc(contextPath+"api/manager/file/rename", handler.RenameFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/file/list", listFileHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/file/copy", copyFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/file/upload", uploadFileHandler).Methods("POST")              // Added upload file handler
	router.HandleFunc(contextPath+"api/manager/file/upload1", uploadLagerFileHandler).Methods("POST", "GET") // Added upload file handler
	router.HandleFunc(contextPath+"api/manager/file/unzip", unzipFileHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/folder/list", listFolderHandler).Methods("GET")
	router.HandleFunc(contextPath+"api/manager/folder/delete", deleteFolderHandler).Methods("DELETE")
	router.HandleFunc(contextPath+"api/manager/folder/rename", renameFolderHandler).Methods("PUT")
	router.HandleFunc(contextPath+"api/manager/folder/copy", copyFolderHandler).Methods("POST")
	router.HandleFunc(contextPath+"api/manager/folder/create", createFolderHandler).Methods("POST")
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
		log.Println("use security mode, user token will be update per day.")
		ticker := time.NewTicker(24 * time.Hour)
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

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {

	file, handler, err := r.FormFile("file")
	path := r.FormValue("path")
	if err != nil {
		response := model.Response{Code: 400, Message: "Failed to get file", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	defer file.Close()
	if path == "." {
		path = root
	}
	f, err := os.OpenFile(path+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to upload file", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	response := model.Response{Code: 200, Message: "File uploaded successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

type FileChunkParam struct {
	//ID               uint64  `json:"id"`
	ChunkNumber      int     `json:"chunkNumber"`
	ChunkSize        float32 `json:"chunkSize"`
	CurrentChunkSize float32 `json:"currentChunkSize"`
	TotalChunks      int     `json:"totalChunks"`
	TotalSize        float64 `json:"totalSize"`
	Identifier       string  `json:"identifier"`
	Filename         string  `json:"filename"`
	RelativePath     string  `json:"relativePath"`
	//Createtime       time.Time      `json:"createtime"`
	//Updatetime       time.Time      `json:"updatetime"`
	File multipart.File `json:"file"`
}

func uploadLagerFileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// 全部默认上传
		response := model.Response{Code: 200, Message: "上传校验", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	} else if r.Method == "POST" {
		// 接收路径
		path := r.FormValue("path")
		// 接收其他参数
		chunkNumber, _ := strconv.Atoi(r.FormValue("chunkNumber"))
		chunkSize, _ := strconv.ParseFloat(r.FormValue("chunkSize"), 32)
		currentChunkSize, _ := strconv.ParseFloat(r.FormValue("currentChunkSize"), 32)
		totalChunks, _ := strconv.Atoi(r.FormValue("totalChunks"))
		totalSize, _ := strconv.ParseFloat(r.FormValue("totalSize"), 32)
		fileChunkParam := FileChunkParam{ChunkNumber: chunkNumber, ChunkSize: float32(chunkSize), CurrentChunkSize: float32(currentChunkSize), TotalChunks: totalChunks,
			TotalSize: totalSize, Identifier: r.FormValue("identifier"), Filename: r.FormValue("filename"), RelativePath: r.FormValue("relativePath")}
		// 接收file
		file, _, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			response := model.Response{Code: 400, Message: "Failed to get file", Data: nil}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonResponse)
			return
		}
		fileChunkParam.File = file
		fullFileName := path + string(os.PathSeparator) + fileChunkParam.Filename
		saveStatus := false
		if fileChunkParam.TotalChunks == 1 {
			saveStatus = uploadSingleFile(fullFileName, fileChunkParam)
		} else {
			saveStatus = uploadFileByRandomAccessFile(fullFileName, fileChunkParam)
		}

		if saveStatus {
			fmt.Println("上传成功")
			response := model.Response{Code: 200, Message: "File uploaded successfully", Data: nil}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
		} else {
			response := model.Response{Code: 500, Message: "Failed to upload file", Data: nil}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonResponse)
			return
		}
	}
}

func uploadSingleFile(resultFileName string, param FileChunkParam) bool {
	saveFile, err := os.Create(resultFileName)
	defer saveFile.Close()
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return false
	}

	_, err = io.Copy(saveFile, param.File)
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return false
	}

	return true
}

func uploadFileByRandomAccessFile(resultFileName string, param FileChunkParam) bool {
	randomAccessFile, err := os.OpenFile(resultFileName, os.O_RDWR|os.O_CREATE, 0666)
	defer randomAccessFile.Close()
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return false
	}

	chunkSize := param.ChunkSize
	if chunkSize == 0 {
		chunkSize = 1024 * 1024
	}
	offset := int64(chunkSize * float32(param.ChunkNumber-1))

	_, err = randomAccessFile.Seek(offset, 0)
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return false
	}

	_, err = io.Copy(randomAccessFile, param.File)
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return false
	}

	return true
}

func deleteFileHandler(w http.ResponseWriter, r *http.Request) {

	filePath := r.FormValue("path")
	if filePath == "" {
		response := model.Response{Code: 400, Message: "Missing path parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	fmt.Println("remove file: " + filePath)
	err := os.Remove(filePath)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to delete file", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "File deleted successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func listFileHandler(w http.ResponseWriter, r *http.Request) {

	files := []model.File{}
	path := r.FormValue("path")
	if len(path) == 0 {
		path = root
	}
	if path == "." {
		path = root
	}
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		response := model.Response{Code: 500, Message: "Failed to list files", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	for _, file := range dir {
		var size int64 = -1
		if !file.IsDir() { // check if it's a file
			size = file.Size()
		}

		files = append(files, model.File{Name: file.Name(), Path: path + "/" + file.Name(), IsDir: file.IsDir(), Size: size, ModTime: file.ModTime()})
		sort.Slice(files, func(i, j int) bool {
			return files[i].IsDir
		})
	}
	response := model.Response{Code: 200, Message: "Files listed successfully", Data: files}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func copyFileHandler(w http.ResponseWriter, r *http.Request) {

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

	// filePath := r.FormValue("path")
	// copyName := r.FormValue("name")
	filePath := requestData.Path
	copyName := requestData.Name

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

func executeCommand(command string, args ...string) error {
	log.Println(command, args)
	cmd := exec.Command(command, args...)

	// 执行命令
	err := cmd.Run()

	return err
}

func unzipFileHandler(w http.ResponseWriter, r *http.Request) {

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
	fileName := requestData.Name

	if filePath == "" || fileName == "" {
		response := model.Response{Code: 400, Message: "Missing path or name parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		return
	}
	// 获取文件所在的目录，并解压到当前目录
	dir := filepath.Dir(filePath)
	var cmdErr error = nil
	if strings.HasSuffix(fileName, ".zip") {
		cmdErr = executeCommand("unzip", filePath, "-d", dir)
	} else if strings.HasSuffix(fileName, ".tar.gz") {
		cmdErr = executeCommand("tar", "-xzf", filePath, "-C", dir)
	} else if strings.HasSuffix(fileName, ".tar") {
		cmdErr = executeCommand("tar", "-xf", filePath, "-C", dir)
	}
	if cmdErr != nil {
		http.Error(w, cmdErr.Error(), http.StatusBadRequest)
		log.Println("unzip failed", cmdErr)
		return
	}

	response := model.Response{Code: 200, Message: "File unzip successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func listFolderHandler(w http.ResponseWriter, r *http.Request) {

	folders := []model.File{}
	path := r.FormValue("path")
	if len(path) == 0 {
		path = root
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
			folders = append(folders, model.File{Name: file.Name(), Path: path + "/" + file.Name(), IsDir: true, Id: strconv.FormatInt(file.ModTime().UnixNano(), 10)})
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
