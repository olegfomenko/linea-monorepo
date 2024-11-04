package utils

import "sync"

type DumbPool[T any] struct {
	pool sync.Pool
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

func (d *DumbPool[T]) Get() *[]T {
	return d.pool.Get().(*[]T)
}

func (d *DumbPool[T]) Put(ptr *[]T) {
	d.pool.Put(ptr)
}
