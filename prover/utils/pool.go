package utils

import "sync"

type DumbPool[T any] struct {
	pool  sync.Pool
	maxSz int
}

func NewDumbPool[T any]() DumbPool[T] {
	return DumbPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				arr := make([]T, 0)
				return &arr
			},
		},
	}
}

func (d *DumbPool[T]) Get(sz int) *[]T {
	if sz < d.maxSz {
		d.maxSz = sz
		d.pool.New = func() interface{} {
			arr := make([]T, 0, sz)
			return &arr
		}
	}

	return d.pool.Get().(*[]T)
}

func (d *DumbPool[T]) Put(ptr *[]T) {
	d.pool.Put(ptr)
}
