package lua

import (
	"context"
)

func (ls *LState) WithValue(key interface{}, v interface{}) context.Context {
	if ls.ctx == nil {
		ls.ctx = context.WithValue(ls.ctx, key, v)
		return ls.ctx
	}

	return context.WithValue(ls.ctx, key, v)
}

func (ls *LState) Value(key interface{}) interface{} {
	if ls.ctx == nil {
		return nil
	}

	return ls.ctx.Value(key)
}

func (ls *LState) SetValue(key interface{}, v interface{}) {
	if ls.ctx == nil {
		ls.ctx = context.WithValue(ls.ctx, key, v)
	}

	ls.ctx = context.WithValue(ls.ctx, key, v)
}
