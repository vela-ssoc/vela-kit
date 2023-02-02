package tasktree

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"time"
)

type ExData struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	State string `json:"state"`
}

type View struct {
	ID      string    `json:"id"`
	Dialect bool      `json:"dialect"`
	Key     string    `json:"key"`
	Link    string    `json:"link"`
	Status  string    `json:"status"`
	Hash    string    `json:"hash"`
	From    string    `json:"way"'`
	Uptime  time.Time `json:"time"`
	Error   error     `json:"err"`
	Proc    []ExData  `json:"vela"`
}

func ToView() []View {
	root.rLock()
	defer root.rUnlock()

	n := root.Len()
	v := make([]View, root.Len())

	for i := 0; i < n; i++ {
		cd := root.CodeVM(i)
		v[i].Key = cd.Key()
		v[i].Link = cd.Link()
		v[i].Status = cd.Status()
		v[i].Hash = cd.Hash()
		v[i].From = cd.From()
		v[i].Uptime = cd.header.uptime
		v[i].Error = cd.Wrap()
		v[i].ID = cd.header.id
		v[i].Dialect = cd.header.dialect

		cd.foreach(func(name string, ud *lua.VelaData) bool {
			v[i].Proc = append(v[i].Proc, ExData{
				Name:  name,
				Type:  ud.Data.Type(),
				State: ud.Data.State().String(),
			})
			return true
		})
	}
	return v
}
