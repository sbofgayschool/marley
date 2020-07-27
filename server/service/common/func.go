package common

import (
	"strings"
)

func GetIdVodId(id string) []string {
	return strings.Split(id, "-")
}
