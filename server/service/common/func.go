package common

import (
	"strings"
)

func GetIdVodId(id string) (string, string) {
	res := strings.SplitN(id, "-", 2)
	if len(res) == 1 {
		return id, ""
	}
	return res[0], res[1]
}
