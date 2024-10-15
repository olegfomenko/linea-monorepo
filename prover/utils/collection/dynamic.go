package collection

import (
	"github.com/consensys/linea-monorepo/prover/utils"
	"math/rand"
)

type node[V any] struct {
	Priority    int64
	Size        uint
	Value       V
	Left, Right *node[V]
}

type DynVec[V any] struct {
	root *node[V]
	sz   uint
}

func NewDynVec[V any]() DynVec[V] {
	return DynVec[V]{}
}

func (v *DynVec[V]) Size() uint {
	return v.sz
}

func (v *DynVec[V]) ListAll() []V {
	res := make([]V, 0, v.sz)
	list(v.root, &res)
	return res
}

func (v *DynVec[V]) Append(value V) uint {
	toInsert := &node[V]{
		Priority: rand.Int63(),
		Size:     1,
		Value:    value,
	}

	if v.sz == 0 {
		v.root = toInsert
		v.sz = 1
		return 1
	}

	v1, v2 := split(v.root, v.sz)
	v.root = merge(v1, merge(v2, toInsert))
	v.sz++

	return v.sz
}

func (v *DynVec[V]) MustRemove(i uint) {
	if i > v.sz {
		utils.Panic("Index does not exist in DynVec")
	}

	treeLeftWithPos, treeRight := split(v.root, i)
	treeLeftWithoutPos, requestedToRemove := split(treeLeftWithPos, i-1)
	if requestedToRemove == nil {
		utils.Panic("Failed to remove from DynVec")
	}

	v.root = merge(treeLeftWithoutPos, treeRight)
	v.sz--
}

func (v *DynVec[V]) MustGet(i uint) V {
	if i > v.sz {
		utils.Panic("Index does not exist in DynVec")
	}

	return get(v.root, i)
}

func get[V any](v *node[V], i uint) V {
	if v == nil {
		utils.Panic("Failed to get from DynVec")
	}

	if size(v.Left)+1 == i {
		return v.Value
	}

	if size(v.Left)+1 < i {
		return get(v.Right, i-size(v.Left)-1)
	}

	return get(v.Left, i)
}

func split[V any](v *node[V], k uint) (*node[V], *node[V]) {
	if v == nil {
		return nil, nil
	}

	if size(v.Left)+1 <= k {
		v1, v2 := split(v.Right, k-size(v.Left)-1)
		v.Right = v1
		update(v)
		return v, v2
	}

	v1, v2 := split(v.Left, k)
	v.Left = v2
	update(v)
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
		update(v1)
		return v1
	}

	v2.Left = merge(v1, v2.Left)
	update(v2)
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

func update[V any](v *node[V]) {
	if v != nil {
		v.Size = size(v.Left) + size(v.Right) + 1
	}
}

func size[V any](v *node[V]) uint {
	if v == nil {
		return 0
	}
	return v.Size
}
