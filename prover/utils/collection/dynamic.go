package collection

import (
	"github.com/consensys/linea-monorepo/prover/utils"
	"math/rand"
)

type node[V any] struct {
	Index       int
	Priority    int64
	Value       V
	Left, Right *node[V]
}

type DynVec[V any] struct {
	root    *node[V]
	sz      int
	counter int
}

func NewDynVec[V any]() DynVec[V] {
	return DynVec[V]{}
}

func (v *DynVec[V]) Size() int {
	return v.sz
}

func (v *DynVec[V]) ListAll() []V {
	res := make([]V, 0, v.sz)
	list(v.root, &res)
	return res
}

func (v *DynVec[V]) Append(value V) int {
	toInsert := &node[V]{
		Priority: rand.Int63(),
		Value:    value,
		Index:    v.counter, // this value is unique and only increments
	}

	v.root = merge(v.root, toInsert)
	v.sz++
	v.counter++

	return toInsert.Index
}

func (v *DynVec[V]) MustRemove(i int) {
	treeLeftWithValue, treeRight := split(v.root, i)
	treeLeftWithoutValue, valueNode := split(treeLeftWithValue, i-1)
	if valueNode == nil {
		utils.Panic("Index %d does not exist in DynVec (len = %d)", i, v.sz)
	}

	v.root = merge(treeLeftWithoutValue, treeRight)
	v.sz--
}

func (v *DynVec[V]) MustGet(i int) V {
	if i > v.sz {
		utils.Panic("Index %d does not exist in DynVec (len = %d)", i, v.sz)
	}

	return get(v.root, i)
}

func get[V any](v *node[V], i int) V {
	if v == nil {
		utils.Panic("Failed to get %d from DynVec", i)
	}

	if v.Index == i {
		return v.Value
	}

	if v.Index < i {
		return get(v.Right, i)
	}

	return get(v.Left, i)
}

func split[V any](v *node[V], k int) (*node[V], *node[V]) {
	if v == nil {
		return nil, nil
	}

	if v.Index <= k {
		v1, v2 := split(v.Right, k)
		v.Right = v1
		return v, v2
	}

	v1, v2 := split(v.Left, k)
	v.Left = v2
	return v1, v
}

func merge[V any](v1, v2 *node[V]) *node[V] {
	if v1 == nil {
		return v2
	}

	if v2 == nil {
		return v1
	}

	if v1.Priority > v2.Priority {
		v1.Right = merge(v1.Right, v2)
		return v1
	}

	v2.Left = merge(v1, v2.Left)
	return v2
}

func list[V any](v *node[V], res *[]V) {
	if v == nil {
		return
	}

	list(v.Left, res)
	*res = append(*res, v.Value)
	list(v.Right, res)
}
