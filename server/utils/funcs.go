package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"time"
)

func RandomString() string {
	res, _ := uuid.NewRandom()
	return res.String()
}

func UnixMillion() int64 {
	return time.Now().UnixNano() / 1000000
}

func GetHash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
