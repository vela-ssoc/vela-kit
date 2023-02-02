package pipe

import (
	"github.com/vela-ssoc/vela-kit/vela"
)

func Seek(n int) func(*Px) {
	return func(px *Px) {
		if n < 0 {
			return
		}
		px.seek = n
	}
}

func Env(env vela.Environment) func(*Px) {
	return func(px *Px) {
		px.xEnv = env
	}
}
