package codec

import "github.com/bytedance/sonic"

type Sonic struct{}

func (s Sonic) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func (s Sonic) Unmarshal(b []byte, v interface{}) error {
	return sonic.Unmarshal(b, v)
}

func (s Sonic) Name() string {
	return "sonic"
}
