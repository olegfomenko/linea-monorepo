package collection

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynVecBasic(t *testing.T) {
	vec := NewDynVec[int]()
	assert.Equal(t, vec.Size(), uint(0))
	assert.Equal(t, vec.Append(10), uint(1))
	assert.Equal(t, vec.Size(), uint(1))
	assert.Equal(t, vec.Append(20), uint(2))
	assert.Equal(t, vec.Size(), uint(2))
	assert.Equal(t, vec.Append(30), uint(3))
	assert.Equal(t, vec.Size(), uint(3))
	assert.Equal(t, vec.Append(10), uint(4))
	assert.Equal(t, vec.Size(), uint(4))
	assert.Equal(t, vec.Append(55), uint(5))
	assert.Equal(t, vec.Size(), uint(5))

	assert.Equal(t, vec.ListAll(), []int{10, 20, 30, 10, 55})

	assert.Equal(t, vec.MustGet(1), 10)
	assert.Equal(t, vec.MustGet(2), 20)
	assert.Equal(t, vec.MustGet(3), 30)
	assert.Equal(t, vec.MustGet(4), 10)
	assert.Equal(t, vec.MustGet(5), 55)

	vec.MustRemove(3)
	assert.Equal(t, vec.MustGet(3), 10)
	assert.Equal(t, vec.Size(), uint(4))

	vec.MustRemove(3)
	assert.Equal(t, vec.MustGet(3), 55)
	assert.Equal(t, vec.Size(), uint(3))

	vec.MustRemove(1)
	assert.Equal(t, vec.MustGet(1), 20)
	assert.Equal(t, vec.Size(), uint(2))

	vec.MustRemove(1)
	assert.Equal(t, vec.MustGet(1), 55)
	assert.Equal(t, vec.Size(), uint(1))

	vec.MustRemove(1)
	assert.Equal(t, vec.Size(), uint(0))
}
