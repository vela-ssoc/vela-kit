package minion

import "github.com/vela-ssoc/vela-kit/kind"

func toBulkDoc(index string, data []byte) []byte {
	enc := kind.NewJsonEncoder()
	enc.WriteByte('{')
	enc.Tab("index")
	enc.KV("_index", index)
	enc.KV("_type", "_doc")
	enc.End("}}")
	enc.Char('\n')
	enc.Copy(data)
	enc.Char('\n')
	return enc.Bytes()
}
