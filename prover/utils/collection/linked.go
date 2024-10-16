package collection

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type list[V any] struct {
	value      V
	prev, next *list[V]
}

type LinkedSet[K comparable] struct {
	start, end *list[K]
	index      map[K]*list[K]
}

func NewLinkedSet[K comparable]() LinkedSet[K] {
	return LinkedSet[K]{
		index: make(map[K]*list[K]),
	}
}

func (l *LinkedSet[K]) Size() int {
	return len(l.index)
}

func (l *LinkedSet[K]) MustAppend(key K) {
	if _, ok := l.index[key]; ok {
		utils.Panic("Key %v already exist in map", key)
	}

	node := &list[K]{
		value: key,
		prev:  l.end,
	}

	// sz > 0
	if l.end != nil {
		l.end.next = node
	}

	// sz = 0
	if l.start == nil {
		l.start = node
	}

	l.end = node
	l.index[key] = node
}

func (l *LinkedSet[K]) MustRemove(key K) {
	var node *list[K]
	var ok bool

	if node, ok = l.index[key]; !ok {
		utils.Panic("Key %v does not exist in map", key)
	}

	defer func() {
		delete(l.index, key)
	}()

	// sz = 1, node = start = end
	if l.start == l.end {
		l.start = nil
		l.end = nil
		return
	}

	if l.start == node {
		l.start = node.next
		l.start.prev = nil
		return
	}

	if l.end == node {
		l.end = node.prev
		l.end.next = nil
		return
	}

	node.prev.next = node.next
	node.next.prev = node.prev
	return
}

func (l *LinkedSet[K]) Exists(ks ...K) bool {
	for _, k := range ks {
		_, found := l.index[k]
		if !found {
			return false
		}
	}
	return true
}

func (l *LinkedSet[K]) MustExists(keys ...K) {
	var missingListString error
	ok := true

	for _, key := range keys {
		if _, found := l.index[key]; !found {

			// accumulate the keys in an user-friendly error message
			if missingListString == nil {
				missingListString = fmt.Errorf("%v", key)
			} else {
				missingListString = fmt.Errorf("%v, %v", missingListString, key)
			}

			ok = false
		}
	}

	if !ok {
		utils.Panic("MustExists : assertion failed. (%v are missing)", missingListString)
	}
}

func (l *LinkedSet[K]) ListAll() []K {
	res := make([]K, 0, len(l.index))
	for node := l.start; node != nil; node = node.next {
		res = append(res, node.value)
	}
	return res
}
