package env

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/audit"
)

func (env *Environment) newAudit() {
	if env.adt != nil {
		return
	}
	env.adt = audit.New()
	err := env.adt.Start()
	if err != nil {
		fmt.Printf("%s audit start fail %v\n", env.Name(), err)
		return
	}

	fmt.Printf("%s audit start succeed\n", env.Name())
}

func (env *Environment) Adt() interface{} {
	return env.adt
}
