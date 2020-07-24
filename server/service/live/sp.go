package live

import (
	"errors"
	"github.com/sbofgayschool/marley/server/utils"
)

type Operation struct {
	Opt         string
	ElapsedTime int64
}

func addOperation(id string, timestamp int64, opt string) error {
	b, ok := broadcasters[id]
	if !ok || b.Timestamp != timestamp {
		return errors.New("broadcasting has stopped")
	}
	b.operations = append(b.operations, &Operation{Opt: opt, ElapsedTime: utils.UnixMillion()})
	return nil
}

func fetchOperations(b *Broadcaster) (res []string) {
	for _, opt := range b.operations {
		res = append(res, opt.Opt)
	}
	return
}
