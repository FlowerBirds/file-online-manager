package util

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GenToken(tokenLength int) string {
	tokenBytes := make([]byte, tokenLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		log.Println("随机生成token失败:", err)
		return ""
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)[:tokenLength]
	return token
}
