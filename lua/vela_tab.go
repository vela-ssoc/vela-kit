package lua

import "strings"

// table check int or default
func (tb *LTable) CheckInt(key string, d int) int {

	v := tb.RawGetString(key)
	switch n := v.(type) {
	case LNumber:
		return int(n)
	default:
		return d
	}
}

// table check int or default
func (tb *LTable) CheckUint32(key string, d uint32) uint32 {

	v := tb.RawGetString(key)
	switch n := v.(type) {
	case LNumber:
		return uint32(n)
	default:
		return d
	}
}

// table check string or default
func (tb *LTable) CheckString(key string, d string) string {
	v := tb.RawGetString(key)
	switch ud := v.(type) {
	case LString:
		return ud.String()
	default:
		return d
	}
}

func (tb *LTable) CheckSocket(key string, L *LState) string {
	v := tb.RawGetString(key).String()
	if e := CheckSocket(v); e != nil {
		L.RaiseError("expected socket invalid  err: %v", e)
		return ""
	}
	return v
}

func (tb *LTable) CheckSockets(key string, L *LState) string {
	v := tb.RawGetString(key).String()
	arr := strings.Split(v, ",")
	var err error

	for _, item := range arr {
		err = CheckSocket(item)
		if err != nil {
			L.RaiseError("expected %s socket error: %v ", item, err)
			return ""
		}
	}
	return v
}

func (tb *LTable) CheckVelaData(L *LState, key string) *VelaData {
	data := tb.RawGetString(key)
	if data.Type() != LTVelaData {
		L.RaiseError("invalid type , %s must be userdata , got %s", key, data.Type().String())
		return nil
	}

	return data.(*VelaData)
}

func (tb *LTable) CheckBool(key string, d bool) bool {
	v := tb.RawGetString(key)
	if lv, ok := v.(LBool); ok {
		return bool(lv)
	}
	return d
}
