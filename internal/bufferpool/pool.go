package bufferpool

import (
	"bytes"
	"sync"
)

type buffPool struct {
	pool sync.Pool
}

func New() *buffPool {
	var bp buffPool
	bp.pool.New = new
	return &bp
}

func new() interface{} {
	return &bytes.Buffer{}
}

func (b *buffPool) Get() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

func (b *buffPool) Release(buf *bytes.Buffer) {
	buf.Reset()
	b.pool.Put(buf)
}
