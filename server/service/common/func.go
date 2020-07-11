package common

import (
    "strings"

    "github.com/sbofgayschool/marley/server/infra/sock"
)

func GetIdVodId(id string) []string {
    return strings.Split(id, "-")
}

func GetCurrentAudience(id string) int {
    return sock.CountGroupClient(id)
}
