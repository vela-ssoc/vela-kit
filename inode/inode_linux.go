package inode

import (
	"github.com/elastic/gosigar"
	"os"
	"strconv"
	"strings"
)

func (inode *Inodes) Value(pid int, link string) {

	inode.Total++
	if !strings.HasPrefix(link, "socket:[") {
		return
	}
	inode.SocketTotal++

	node, err := strconv.ParseInt(link[8:len(link)-1], 10, 64)
	if err != nil {
		return
	}
	inode.socket[uint32(node)] = pid
}

func (inode *Inodes) read(pid int) {
	path := "/proc" + "/" + strconv.Itoa(pid) + "/fd/"
	d, err := os.Open(path)
	if err != nil {
		return
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return
	}

	for _, name := range names {
		pathLink := path + name
		target, er := os.Readlink(pathLink)
		if er != nil {
			continue
		}
		inode.Value(pid, target)
	}

	inode.collect[pid] = len(names)
}

func (inode *Inodes) List() []int {
	if len(inode.list) > 0 {
		return inode.list
	}

	pl := gosigar.ProcList{}
	er := pl.Get()
	if er != nil {
		return nil
	}

	return pl.List
}

func (inode *Inodes) R() error {
	inode.socket = make(map[uint32]int)
	inode.collect = make(map[int]int)
	list := inode.List()
	n := len(list)
	for i := 0; i < n; i++ {
		inode.read(list[i])
	}

	return nil
}

func (inode *Inodes) FindPid(id uint32) int {
	return inode.socket[id]
}
