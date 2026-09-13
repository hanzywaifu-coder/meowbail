package protocol

import (
	"bytes"
	"sync"
)

// BufferPool menyediakan reusable *bytes.Buffer untuk streaming media, decoding JSON, dan formatting string
// Memangkas alokasi heap dan meringankan beban GC saat load tinggi
var (
	SmallBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4*1024)) // 4KB
		},
	}
	MediumBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 64*1024)) // 64KB
		},
	}
	LargeBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 512*1024)) // 512KB
		},
	}
)

// GetSmallBuffer mengambil buffer 4KB dari pool
func GetSmallBuffer() *bytes.Buffer {
	buf := SmallBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutSmallBuffer mengembalikan buffer 4KB ke pool
func PutSmallBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 16*1024 {
		buf.Reset()
		SmallBufferPool.Put(buf)
	}
}

// GetMediumBuffer mengambil buffer 64KB dari pool
func GetMediumBuffer() *bytes.Buffer {
	buf := MediumBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutMediumBuffer mengembalikan buffer 64KB ke pool
func PutMediumBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 256*1024 {
		buf.Reset()
		MediumBufferPool.Put(buf)
	}
}

// GetLargeBuffer mengambil buffer 512KB dari pool
func GetLargeBuffer() *bytes.Buffer {
	buf := LargeBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutLargeBuffer mengembalikan buffer 512KB ke pool
func PutLargeBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 2*1024*1024 {
		buf.Reset()
		LargeBufferPool.Put(buf)
	}
}
