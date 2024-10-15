package collection

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynVecBasic(t *testing.T) {
	vec := NewDynVec[int]()
	assert.Equal(t, vec.Size(), 0)
	assert.Equal(t, vec.Append(10), 0)
	assert.Equal(t, vec.Size(), 1)
	assert.Equal(t, vec.Append(20), 1)
	assert.Equal(t, vec.Size(), 2)
	assert.Equal(t, vec.Append(30), 2)
	assert.Equal(t, vec.Size(), 3)
	assert.Equal(t, vec.Append(10), 3)
	assert.Equal(t, vec.Size(), 4)
	assert.Equal(t, vec.Append(55), 4)
	assert.Equal(t, vec.Size(), 5)

	assert.Equal(t, vec.ListAll(), []int{10, 20, 30, 10, 55})

	assert.Equal(t, vec.MustGet(0), 10)
	assert.Equal(t, vec.MustGet(1), 20)
	assert.Equal(t, vec.MustGet(2), 30)
	assert.Equal(t, vec.MustGet(3), 10)
	assert.Equal(t, vec.MustGet(4), 55)

	vec.MustRemove(3)
	assert.Equal(t, vec.Size(), (4))
	vec.MustRemove(1)
	assert.Equal(t, vec.Size(), (3))
	vec.MustRemove(2)
	assert.Equal(t, vec.Size(), (2))
	vec.MustRemove(4)
	assert.Equal(t, vec.Size(), (1))
	vec.MustRemove(0)
	assert.Equal(t, vec.Size(), (0))

	assert.Equal(t, vec.Append(100), (5))
	assert.Equal(t, vec.Size(), 1)
}
