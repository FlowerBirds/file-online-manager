package model

import (
	"encoding/json"
	"time"
)

type File struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Id      string    `json:"id"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type RequestFileData struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func (f *File) MarshalJSON() ([]byte, error) {
	type Alias File // 创建一个别名类型，以便访问原始 File 结构体的字段

	return json.Marshal(&struct {
		*Alias
		ModTime string `json:"modTime"`
	}{
		Alias:   (*Alias)(f),
		ModTime: f.ModTime.Format("2006-01-02 15:04:05"),
	})
}
