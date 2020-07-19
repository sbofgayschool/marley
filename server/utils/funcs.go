package utils

import (
	"github.com/google/uuid"
)

func RandomString() string {
	res, _ := uuid.NewRandom()
	return res.String()
}
