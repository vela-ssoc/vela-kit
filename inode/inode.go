package inode

type Inodes struct {
	list        []int
	Total       int64
	SocketTotal int64
	socket      map[uint32]int //socket inode InodeMap
	collect     map[int]int    //pid num inode number
}

func New(v []int) *Inodes {
	inode := &Inodes{list: v}
	inode.R()
	return inode
}

func All() *Inodes {
	inode := &Inodes{list: nil}
	inode.R()
	return inode
}
