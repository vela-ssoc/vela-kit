package env

import (
	"github.com/asdine/storm/v3"
	"github.com/vela-ssoc/vela-kit/bucket"
	"github.com/vela-ssoc/vela-kit/codec"
	"github.com/vela-ssoc/vela-kit/execpt"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.etcd.io/bbolt"
	"path/filepath"
)

type bboltDB struct {
	path string
	db   *bbolt.DB
	db2  *storm.DB
}

func (env *Environment) openDb() {

	//发现文件
	path := filepath.Join(env.ExecDir(), ".ssc.db")

	opt := &bbolt.Options{
		Timeout:      0,
		NoGrowSync:   false,
		NoSync:       true,
		FreelistType: bbolt.FreelistMapType,
	}
	//新建数据存储
	db, err := bbolt.Open(path, 0600, opt)
	execpt.Fatal(err)

	db2, err := storm.Open(".ssc.db", storm.UseDB(db))
	db2.WithCodec(codec.Sonic{})
	execpt.Fatal(err)

	env.bdb = &bboltDB{
		path: path,
		db:   db,
		db2:  db2,
	}
}

func (env *Environment) Bucket(v ...string) vela.Bucket {
	return bucket.Pack(env, v...)
}

func (env *Environment) Storm(v ...string) storm.Node {
	return env.bdb.db2.From(v...)
}
