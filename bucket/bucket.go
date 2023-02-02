package bucket

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.etcd.io/bbolt"
)

type Tx = bbolt.Tx

type Bucket struct {
	db     *bbolt.DB
	chains [][]byte
	export string
}

func (bkt *Bucket) NewExpireQueue() *expireQueue {
	return &expireQueue{bkt: bkt}
}

func Pack(env vela.Environment, names ...string) *Bucket {
	var chains [][]byte

	for _, name := range names {
		chains = append(chains, lua.S2B(name))
	}

	return &Bucket{
		db:     env.DB(),
		chains: chains,
	}
}
