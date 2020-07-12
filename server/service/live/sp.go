package live

import (
	"errors"
	"time"
)

type operation struct {
	opt         string
	elapsedTime int64
}

func addOperation(id string, timestamp int64, opt string) error {
	b, ok := broadcasters[id]
	if !ok || b.Timestamp != timestamp {
		return errors.New("broadcasting has stopped")
	}
	b.operations = append(b.operations, &operation{opt: opt, elapsedTime: time.Now().Unix() - b.Timestamp})
	return nil
}

func fetchOperations(b *Broadcaster) ([]string, error) {
	var res []string
	for _, opt := range b.operations {
		res = append(res, opt.opt)
	}
	return res, nil
}
