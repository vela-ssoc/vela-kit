package bucket

import (
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/vela"
)

type expireQueue struct {
	bkt  *Bucket
	data []string
}

func (ee *expireQueue) IsExpire(key string, it item) {
	if it.mime != vela.EXPIRE {
		return
	}
	ee.data = append(ee.data, key)
}

func (ee *expireQueue) clear() {
	if ee.bkt.db == nil {
		return
	}

	if len(ee.data) == 0 {
		return
	}

	tx, err := ee.bkt.db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	bbt, err := ee.bkt.unpack(tx, false)
	if err != nil {
		return
	}

	for _, val := range ee.data {
		bbt.Delete(auxlib.S2B(val))
	}

	err = tx.Commit()
	if err != nil {
		xEnv.Errorf("%s expire clear fail %v", ee.bkt.Names(), err)
	}
}
