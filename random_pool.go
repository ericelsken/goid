package goid

import (
	"crypto/rand"
	"io"
	"sync"
)

var randomSource io.Reader = rand.Reader

func SetSource(r io.Reader) {
	randomSource = r
}

type randomPool struct {
	mu     sync.Mutex
	buffer []byte
	index  int
}

var pool randomPool

func EnableRandomPool() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.buffer == nil {
		pool.buffer = make([]byte, 1024) // 1024 / 16 = 64 ids in the pool.
		pool.index = len(pool.buffer)
	}
}

func DisableRandomPool() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.buffer = nil
	pool.index = 0
}

func usingPool() bool {
	return pool.buffer != nil
}

func (rp *randomPool) next() ([16]byte, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	result := [16]byte{}
	if rp.index == len(rp.buffer) {
		if _, err := io.ReadFull(randomSource, rp.buffer); err != nil {
			return result, err
		}
		rp.index = 0
	}
	copy(result[:], rp.buffer[rp.index:rp.index+16])
	rp.index += 16
	return result, nil
}
