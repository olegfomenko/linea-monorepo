package collection

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// A set is an unordered collection addressed by keys, which supports
type Set[T comparable] struct {
	inner map[T]struct{}
}

// Constructor for KVStore
func NewSet[K comparable]() Set[K] {
	return Set[K]{
		inner: make(map[K]struct{}),
	}
}

// Returns the list of all the keys
func (kv *Set[K]) ListAll() []K {
	var res []K = make([]K, 0, len(kv.inner))
	for k := range kv.inner {
		res = append(res, k)
	}
	return res
}

// Returns `true` if the entry exists
func (kv *Set[K]) Exists(ks ...K) bool {
	for _, k := range ks {
		_, found := kv.inner[k]
		if !found {
			return false
		}
	}
	return true
}

// Panic if the given entry does not exists
func (kv *Set[K]) MustExists(keys ...K) {
	var missingListString error
	ok := true

	for _, key := range keys {
		if _, found := kv.inner[key]; !found {

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

/*
Inserts regarless of whether the entry was already
present of not. Returns whether the entry was present
already.
*/
func (kv *Set[K]) Insert(k K) bool {
	if _, ok := kv.inner[k]; !ok {
		kv.inner[k] = struct{}{}
		return false
	}
	return true
}

// InsertNew inserts a new value and panics if it was
// contained already
func (kv *Set[K]) InsertNew(key K) {
	if kv.Insert(key) {
		utils.Panic("Entry %v already found", key)
	}
}

// Delete an entry. Panic if the entry was not found
func (kv *Set[K]) MustDel(k K) {
	// Sanity-check
	kv.MustExists(k)
	delete(kv.inner, k)
}
