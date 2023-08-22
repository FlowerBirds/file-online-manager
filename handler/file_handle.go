package handler

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"file-online-manager/model"
	"file-online-manager/util"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var RootPath = "."

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {

	filePath := r.FormValue("path")
	if filePath == "" {
		response := model.Response{Code: 400, Message: "Missing path parameter", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}
	fmt.Println("remove file: " + filePath)
	// 使用RemoveAll则删除文件夹，暂时不实现
	err := os.Remove(filePath)
	if err != nil {
		response := model.Response{Code: 500, Message: "Failed to delete file", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}
	response := model.Response{Code: 200, Message: "File deleted successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

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

func ListFileHandler(root string, w http.ResponseWriter, r *http.Request) {

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

		files = append(files, model.File{Name: file.Name(), Path: path + "/" + file.Name(), IsDir: file.IsDir(), Size: size, ModTime: file.ModTime(), Id: uuid.New().String()})
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

func CopyFileHandler(w http.ResponseWriter, r *http.Request) {

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

	util.CopyFile(filePath, newPath)

	response := model.Response{Code: 200, Message: "File copied successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func UploadLagerFileHandler(w http.ResponseWriter, r *http.Request) {

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
		if strings.Index(path, "./") == 0 || path == "." {
			path = RootPath + "/" + path
		}
		// 接收其他参数
		chunkNumber, _ := strconv.Atoi(r.FormValue("chunkNumber"))
		chunkSize, _ := strconv.ParseFloat(r.FormValue("chunkSize"), 32)
		currentChunkSize, _ := strconv.ParseFloat(r.FormValue("currentChunkSize"), 32)
		totalChunks, _ := strconv.Atoi(r.FormValue("totalChunks"))
		totalSize, _ := strconv.ParseFloat(r.FormValue("totalSize"), 32)
		fileChunkParam := model.FileChunkParam{ChunkNumber: chunkNumber, ChunkSize: float32(chunkSize), CurrentChunkSize: float32(currentChunkSize), TotalChunks: totalChunks,
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
			log.Println("上传成功：", fileChunkParam.Filename, chunkNumber, totalChunks)
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

func uploadSingleFile(resultFileName string, param model.FileChunkParam) bool {
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

func uploadFileByRandomAccessFile(resultFileName string, param model.FileChunkParam) bool {
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

func UnzipFileHandler(w http.ResponseWriter, r *http.Request) {

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
		cmdErr = util.ExecuteCommand("unzip", filePath, "-d", dir)
	} else if strings.HasSuffix(fileName, ".tar.gz") {
		cmdErr = util.ExecuteCommand("tar", "-xzf", filePath, "-C", dir)
	} else if strings.HasSuffix(fileName, ".tar") {
		cmdErr = util.ExecuteCommand("tar", "-xf", filePath, "-C", dir)
	}
	if cmdErr != nil {
		response := model.Response{Code: 400, Message: cmdErr.Error(), Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		log.Println("unzip failed", cmdErr)
		return
	}

	response := model.Response{Code: 200, Message: "File unzip successfully", Data: nil}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取请求的文件名
	filename := r.URL.Query().Get("filename")
	filePath := r.URL.Query().Get("path")
	log.Println(filename)
	log.Println(filePath)
	// 根据文件名的后缀判断文件类型
	contentType := "application/octet-stream"

	// 设置响应头，指定文件的Content-Disposition为attachment，表示下载文件
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)

	// 读取文件内容
	file, err := os.Open(path.Join(filePath, filename))
	defer file.Close()
	if err != nil {
		// 处理文件打开失败的情况
		response := model.Response{Code: 400, Message: "Folder already exists", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}

	// 将文件内容写入响应体
	_, err = io.Copy(w, file)
	if err != nil {
		// 处理文件写入响应体失败的情况
		response := model.Response{Code: 400, Message: "Folder not exists", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
}

func ViewZipFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	log.Println("view zip file:", filePath)
	if !strings.HasSuffix(filePath, ".zip") {
		response := model.Response{Code: 400, Message: "File type unsupported", Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		response := model.Response{Code: 400, Message: err.Error(), Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	defer zipReader.Close()
	fileNames := make([]model.File, 0)
	// 遍历ZIP文件中的文件列表
	for _, file := range zipReader.File {
		// 输出文件名
		// fmt.Println(file.Name)
		fileNames = append(fileNames, model.File{
			Name:    file.Name,
			Size:    int64(file.CompressedSize64),
			ModTime: file.Modified,
		})
	}

	response := model.Response{Code: 200, Message: "File view successfully", Data: fileNames}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func ReleaseZipFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	log.Println("release zip file:", filePath)
	if !strings.HasSuffix(filePath, ".zip") {
		util.Error(w, errors.New("file type unsupported"))
		return
	}
	dir := filepath.Dir(filePath)
	// 提取deleted.conf文件
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		util.Error(w, err)
		return
	}
	defer zipReader.Close()
	deleteFiles := make([]model.File, 0)
	for _, f := range zipReader.File {
		if f.Name == "delete.conf" {
			rc, err := f.Open()
			if err != nil {
				util.Error(w, err)
				return
			}
			defer rc.Close()

			scanner := bufio.NewScanner(rc)
			content := make([]string, 0)
			for scanner.Scan() {
				content = append(content, scanner.Text())
			}
			if err = scanner.Err(); err != nil {
				util.Error(w, err)
				return
			}
			for _, dFile := range content {
				if strings.TrimSpace(dFile) == "" {
					continue
				}
				deleteFile := filepath.Join(dir, dFile)
				log.Println("delete file by delete.conf:", deleteFile)
				fs, err := os.Stat(deleteFile)
				if err != nil {
					log.Printf(err.Error(), "with", dFile)
					deleteFiles = append(deleteFiles, model.File{
						Name: dFile,
						Size: -1,
					})
				} else {
					deleteFiles = append(deleteFiles, model.File{
						Name:    dFile,
						Size:    fs.Size(),
						ModTime: fs.ModTime(),
					})
				}
				os.RemoveAll(deleteFile)
			}
			// return
		}
	}

	// 执行解压操作
	cmdErr := util.ExecuteCommand("unzip", "-o", filePath, "-d", dir)
	if cmdErr != nil {
		response := model.Response{Code: 500, Message: cmdErr.Error(), Data: nil}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonResponse)
		log.Println("release zip failed", cmdErr)
		return
	}

	response := model.Response{Code: 200, Message: "Patch file unzip successfully", Data: deleteFiles}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
