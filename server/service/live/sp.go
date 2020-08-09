package live

import (
	"errors"
	"github.com/sbofgayschool/marley/server/service/common"
	"github.com/sbofgayschool/marley/server/utils"
)

func addOperation(id string, timestamp int64, opt string) error {
	lock.RLock()
	b, ok := broadcasters[id]
	lock.RUnlock()
	if !ok || b.Timestamp != timestamp {
		return errors.New("broadcasting has stopped")
	}
	b.operations = append(b.operations, &common.Operation{Opt: opt, ElapsedTime: utils.UnixMillion()})
	return nil
}

func fetchOperations(b *Broadcaster) (res []string) {
	for _, opt := range b.operations {
		res = append(res, opt.Opt)
	}
	return
}
