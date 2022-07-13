package pool

import (
	"sync"
)

var (
	bufPool    sync.Pool
	bufPool1k  sync.Pool
	bufPool2k  sync.Pool
	bufPool4k  sync.Pool
	bufPool8k  sync.Pool
	bufPool16k sync.Pool
)

const (
	k16 = 16 * 1024
	k8  = 8 * 1024
	k4  = 4 * 1024
	k2  = 2 * 1024
	k1  = 1024
)

func GetBuf(size int) []byte {
	var x interface{}
	switch {
	case size >= k16:
		x = bufPool16k.Get()
	case size >= k8:
		x = bufPool8k.Get()
	case size >= k4:
		x = bufPool4k.Get()
	case size >= k2:
		x = bufPool2k.Get()
	case size >= k1:
		x = bufPool1k.Get()
	default:
		x = bufPool.Get()
	}
	if x == nil {
		return make([]byte, size)
	}
	buf := x.([]byte)
	if cap(buf) < size {
		return make([]byte, size)
	}
	return buf[:size]
}

func PutBuf(buf interface{}) {
	size := cap(buf.([]byte))
	switch {
	case size >= k16:
		bufPool16k.Put(buf)
	case size >= k8:
		bufPool8k.Put(buf)
	case size >= k4:
		bufPool4k.Put(buf)

	case size >= k2:
		bufPool2k.Put(buf)

	case size >= k1:
		bufPool1k.Put(buf)
	default:
		bufPool.Put(buf)
	}
}
