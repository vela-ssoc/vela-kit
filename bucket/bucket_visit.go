package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/mime"
	"github.com/vela-ssoc/vela-kit/strutil"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.etcd.io/bbolt"
	"strings"
	"time"
)

func (bkt *Bucket) Type() lua.LValueType                   { return lua.LTObject }
func (bkt *Bucket) AssertFloat64() (float64, bool)         { return 0, false }
func (bkt *Bucket) AssertString() (string, bool)           { return "", false }
func (bkt *Bucket) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (bkt *Bucket) Peek() lua.LValue                       { return bkt }

func (bkt *Bucket) Encode(iv interface{}, expire int) ([]byte, error) {
	it := &element{}
	err := iEncode(it, iv, expire)
	if err != nil {
		return nil, err
	}

	return it.Byte(), nil
}

func (bkt *Bucket) BatchStore(v map[string]interface{}, expire int) error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		if v == nil {
			return nil
		}

		if len(v) == 0 {
			return nil
		}

		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return err
		}

		for key, iv := range v {
			it := element{}
			err = iEncode(&it, iv, expire)
			if err != nil {
				return err
			}
			bbt.Put(auxlib.S2B(key), it.Byte())
		}
		return nil
	})

}

func (bkt *Bucket) Store(key string, v interface{}, expire int) error {
	err := bkt.dbx.Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return err
		}
		var it element
		err = iEncode(&it, v, expire)
		if err != nil {
			return err
		}
		bbt.Put(auxlib.S2B(key), it.Byte())
		return nil
	})
	return err
}

func (bkt *Bucket) Load(key string) (element, error) {
	elem := element{}
	od := bkt.NewOverdue()

	err := bkt.dbx.View(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, true)
		if err != nil {
			return err
		}

		data := bbt.Get(auxlib.S2B(key))
		err = iDecode(&elem, data)
		return err
	})

	od.IsExpire(key, elem)
	od.clear()

	if err != nil {
		return elem, err
	}

	return elem, nil
}

func (bkt *Bucket) Replace(key string, v interface{}, expire int) error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return err
		}
		kb := auxlib.S2B(key)
		data := bbt.Get(kb)

		chunk, name, err := mime.Encode(v)
		if err != nil {
			return err
		}

		var it element
		err = iDecode(&it, data)
		if err != nil {
			it.set(name, chunk, expire)
			return err
		}

		it.mime = name
		it.chunk = chunk
		return bbt.Put(kb, it.Byte())
	})
}

func (bkt *Bucket) Delete(key string) error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return nil
		}
		return bbt.Delete(auxlib.S2B(key))
	})
}

func (bkt *Bucket) Ext() string {
	n := len(bkt.chains)
	if n == 0 {
		return ""
	}

	return string(bkt.chains[n-1])
}

func (bkt *Bucket) Clean() error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		n := len(bkt.chains)
		switch n {
		case 0:
			return nil
		case 1:
			return tx.DeleteBucket(bkt.chains[0])
		default:
			ext := bkt.chains[n-1]
			bkt.chains = bkt.chains[:n-1]

			bbt, er := bkt.unpack(tx, false)
			if er != nil {
				goto done
			}

			er = bbt.DeleteBucket(ext)
			if er != nil {
				goto done
			}

			_, er = bbt.CreateBucketIfNotExists(ext)
		done:
			bkt.chains = append(bkt.chains, ext)
			return er
		}
	})
}

func (bkt *Bucket) DeleteBucket(nb string) error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return nil
		}
		return bbt.DeleteBucket(auxlib.S2B(nb))
	})
}

func (bkt *Bucket) Incr(key string, val int, expire int) (int, error) {
	var sum float64

	err := xEnv.DB().Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return err
		}

		b := lua.S2B(key)
		data := bbt.Get(b)
		it := &element{}
		err = iDecode(it, data)
		if err != nil {
			xEnv.Infof("incr %s decode fail error %v", key, err)
			goto INCR
		}

	INCR:
		sum = it.incr(float64(val), expire)
		return bbt.Put(b, it.Byte())
	})

	if err != nil {
		return int(sum), err
	}

	return int(sum), nil
}

func (bkt *Bucket) Get(key string) (interface{}, error) {
	it, err := bkt.Load(key)
	if err != nil {
		return nil, err
	}
	return it.Decode()
}

func (bkt *Bucket) GetFn(key string, fn func(interface{}) error) error {
	it, err := bkt.Load(key)
	if err != nil {
		return err
	}

	src, err := it.Decode()
	if err != nil {
		return err
	}

	return fn(src)
}

func (bkt *Bucket) Int(key string) int {
	val, err := bkt.Get(key)
	if err != nil {
		return 0
	}
	return auxlib.ToInt(val)
}

func (bkt *Bucket) Int64(key string) int64 {
	val, err := bkt.Get(key)
	if err != nil {
		return 0
	}
	return auxlib.ToInt64(val)
}

func (bkt *Bucket) Bool(key string) bool {
	val, err := bkt.Get(key)
	if err != nil {
		return false
	}

	return auxlib.ToBool(val)
}

func (bkt *Bucket) unpack(tx *Tx, readonly bool) (*bbolt.Bucket, error) {
	var bbt *bbolt.Bucket
	var err error

	n := len(bkt.chains)
	if n < 1 {
		return nil, errors.New("not found bucket")
	}

	bbt, err = Tx2B(tx, bkt.chains[0], readonly)
	if n == 1 {
		return bbt, err
	}

	//如果报错
	if err != nil {
		return bbt, err
	}

	for i := 1; i < n; i++ {
		bbt, err = Bkt2B(bbt, bkt.chains[i], readonly)
		if err != nil {
			return nil, err
		}
	}

	return bbt, nil
}
func (bkt *Bucket) Batch(fn func(tx *bbolt.Tx, bbt *bbolt.Bucket) error, writable bool) error {
	if fn == nil {
		return fmt.Errorf("not found found")
	}

	if bkt.dbx == nil {
		return fmt.Errorf("not found db")
	}

	tx, err := bkt.dbx.Begin(writable)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bbt, err := bkt.unpack(tx, !writable)
	if err != nil {
		return err
	}

	fn(tx, bbt)
	return tx.Commit()
}

func (bkt *Bucket) Push(key string, val []byte, expire int64) error {
	return bkt.dbx.Batch(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, false)
		if err != nil {
			return err
		}

		var ttl uint64
		if expire <= 0 {
			ttl = 0
		} else {
			ttl = uint64(time.Now().Unix() + expire)
		}

		it := element{
			mime:  vela.BYTES,
			size:  uint64(len(vela.BYTES)),
			ttl:   ttl,
			chunk: val,
		}
		return bbt.Put(auxlib.S2B(key), it.Byte())
	})
}

func (bkt *Bucket) Value(key string) ([]byte, error) {
	od := bkt.NewOverdue()
	it, err := bkt.Load(key)
	if err != nil {
		return nil, err
	}

	switch it.mime {
	case vela.EXPIRE:
		od.IsExpire(key, it)
		return nil, nil

	case vela.NIL:
		return nil, nil
	case vela.STRING, vela.BYTES, "[]int8":
		return it.chunk, nil

	default:
		return nil, fmt.Errorf("%s not bytes , got %s", key, it.mime)
	}
}

func (bkt *Bucket) Count() int {
	var s bbolt.BucketStats

	bkt.dbx.View(func(tx *bbolt.Tx) error {
		bbt, err := bkt.unpack(tx, true)
		if err != nil {
			return err
		}

		s = bbt.Stats()
		return nil
	})
	return s.KeyN
}
func (bkt *Bucket) ForEach(fn func(string, []byte)) {
	od := bkt.NewOverdue()
	err := bkt.dbx.View(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, true)
		if err != nil {
			return err
		}
		err = bbt.ForEach(func(k, data []byte) error {
			var it element

			e := iDecode(&it, data)
			if e != nil {
				xEnv.Infof("%s decode data fail %s", bkt.Names(), string(data))
				return nil
			}
			key := string(k)

			od.IsExpire(key, it)

			switch it.mime {
			case "[]int8": //old mime
				fn(key, it.chunk)
			case vela.STRING, vela.BYTES:
				fn(key, it.chunk)
			default:
				fn(key, nil)
			}
			return nil
		})
		return err
	})

	if err != nil && !strings.HasSuffix(err.Error(), "not found") {
		xEnv.Infof("bucket range fail %v", err)
		return
	}

	od.clear()
}

func (bkt *Bucket) Range(fn func(string, interface{})) {
	od := bkt.NewOverdue()
	err := bkt.dbx.View(func(tx *Tx) error {
		bbt, err := bkt.unpack(tx, true)
		if err != nil {
			return err
		}

		err = bbt.ForEach(func(k, data []byte) error {

			elem := element{}
			err = iDecode(&elem, data)
			if err != nil {
				xEnv.Errorf("bucket decode fail mime:%s size:%d %v", elem.mime, elem.size, err)
				return nil
			}
			key := string(k)
			od.IsExpire(key, elem)

			v, er := elem.Decode()
			if er != nil {
				xEnv.Errorf("bucket decode fail mime:%s size:%d %v", elem.mime, elem.size, er)
				return nil
			}

			fn(key, v)
			return nil
		})

		return err
	})

	if err != nil {
		xEnv.Errorf("bucket range fail %v", err)
		return
	}

	od.clear()
}

func (bkt *Bucket) Json() *Bucket {
	bkt.export = "json"
	return bkt
}

func (bkt *Bucket) ReadOnly() vela.Bucket {
	bkt.readOnly = true
	return bkt
}

func (bkt *Bucket) Names() string {
	return string(bytes.Join(bkt.chains, []byte(",")))
}

func (bkt *Bucket) String() string {
	switch bkt.export {
	case "json":
		return strutil.B2S(Bkt2Json(bkt))
	case "line":
		return strutil.B2S(Bkt2Line(bkt))
	default:
		return fmt.Sprintf("%p", bkt)
	}
}
