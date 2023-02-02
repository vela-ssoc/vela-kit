package bucket

import (
	"bytes"
	"fmt"
	auxlib "github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.etcd.io/bbolt"
)

func Tx2B(tx *Tx, name []byte, readonly bool) (*bbolt.Bucket, error) {
	if readonly {
		bkt := tx.Bucket(name)
		if bkt == nil {
			return nil, fmt.Errorf("%s not found", auxlib.B2S(name))
		}
		return bkt, nil
	}

	return tx.CreateBucketIfNotExists(name)
}

func Bkt2B(b *bbolt.Bucket, name []byte, readonly bool) (*bbolt.Bucket, error) {
	if readonly {
		bkt := b.Bucket(name)
		if bkt == nil {
			return nil, fmt.Errorf("%s not found", auxlib.B2S(name))
		}
		return bkt, nil
	}
	return b.CreateBucketIfNotExists(name)
}

func bucketExportHelper(bkt *Bucket, fn func(string, interface{})) error {
	ee := bkt.NewExpireQueue()
	err := xEnv.DB().View(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, true)
		if err != nil {
			return err
		}

		err = bbt.ForEach(func(k, v []byte) error {
			var it item
			if e := iDecode(&it, v); e != nil {
				xEnv.Errorf("export json %s decode error %v", k, e)
				return nil
			}
			ee.IsExpire(string(k), it)

			switch it.mime {
			case vela.NIL:
				return nil
			case "lua.LNilType":
				return nil
			}

			iv, ie := it.Decode()
			if ie != nil {
				xEnv.Errorf("export json %s to interface error %v", k, ie)
				return nil
			}

			fn(auxlib.B2S(k), iv)
			return nil
		})

		return err
	})

	ee.clear()

	if err != nil {
		return err
	}

	return nil
}

func bucketExportJson(bkt *Bucket) []byte {
	buf := kind.NewJsonEncoder()
	buf.Tab("")
	err := bucketExportHelper(bkt, buf.KV)
	if err != nil {
		xEnv.Errorf("export %v error %v", bkt, err)
		return nil
	}

	if err != nil {
		return nil
	}

	buf.End("}")
	return buf.Bytes()
}

func bucketExportLine(bkt *Bucket) []byte {
	var buf bytes.Buffer
	fn := func(key string, v interface{}) {
		buf.WriteString(key)
		buf.WriteByte(':')
		buf.WriteString(auxlib.ToString(v))
		buf.WriteByte('\n')
	}

	err := bucketExportHelper(bkt, fn)
	if err != nil {
		xEnv.Errorf("export %v error %v", bkt.chains, err)
		return nil
	}

	return buf.Bytes()
}

func decode(v []byte) (interface{}, error) {
	var it item
	err := iDecode(&it, v)
	if err != nil {
		return nil, err
	}

	return it.Decode()
}
