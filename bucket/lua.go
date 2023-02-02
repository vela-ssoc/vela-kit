package bucket

import (
	"github.com/vela-ssoc/vela-kit/execpt"
	lua "github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.etcd.io/bbolt"
)

var xEnv vela.Environment

func checkBD(L *lua.LState) *bbolt.DB {
	if xEnv.DB() == nil {
		xEnv.Error("bbolt db not found")
		L.RaiseError("bbolt db not found")
		return nil
	}

	return xEnv.DB()
}

func newLuaBucket(L *lua.LState) int {
	_ = checkBD(L)

	n := L.GetTop()
	if n == 0 {
		L.Push(lua.LNil)
		return 1
	}

	b := &Bucket{db: xEnv.DB()}

	for i := 1; i <= n; i++ {
		name := L.CheckString(i)
		b.chains = append(b.chains, lua.S2B(name))
	}

	L.Push(b)
	return 1

}

func luaBucketInfoApi(L *lua.LState) int {
	tab := L.NewTable()
	db := checkBD(L)
	i := 0
	db.View(func(tx *Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			i++
			tab.RawSetInt(i, lua.B2L(name))
			return nil
		})
		return nil
	})
	L.Push(tab)
	return 1
}

func luaBucketRemoveApi(L *lua.LState) int {
	db := checkBD(L)

	n := L.GetTop()
	if n == 0 {
		return 0
	}

	err := db.Batch(func(tx *Tx) error {
		errs := execpt.New()
		for i := 1; i <= n; i++ {
			lv := L.Get(i)
			if lv.Type() == lua.LTString {
				name := lv.String()
				errs.Try(name, tx.DeleteBucket(lua.S2B(name)))
			}
		}
		return errs.Wrap()
	})

	if err == nil {
		return 0
	}

	L.Push(lua.S2L(err.Error()))
	return 1
}

func indexL(L *lua.LState, key string) lua.LValue {
	db := checkBD(L)

	return &Bucket{
		db:     db,
		chains: [][]byte{[]byte(key)},
		export: "json",
	}

}

func Constructor(env vela.Environment) {
	xEnv = env
	uv := lua.NewUserKV()
	uv.Set("info", lua.NewFunction(luaBucketInfoApi))
	uv.Set("bucket", lua.NewFunction(newLuaBucket))
	uv.Set("remove", lua.NewFunction(luaBucketRemoveApi))
	xEnv.Set("db",
		lua.NewExport("vela.db.export",
			lua.WithTable(uv),
			lua.WithFunc(newLuaBucket),
			lua.WithIndex(indexL)))
}
