package utils

import (
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
