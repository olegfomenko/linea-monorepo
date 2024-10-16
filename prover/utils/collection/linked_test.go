package collection

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkedSet(t *testing.T) {
	vec := NewLinkedSet[int]()
	assert.Equal(t, vec.Size(), 0)
	vec.MustAppend(1)
	assert.Equal(t, vec.Size(), 1)
	vec.MustAppend(2)
	assert.Equal(t, vec.Size(), 2)
	vec.MustAppend(3)
	assert.Equal(t, vec.Size(), 3)
	vec.MustAppend(4)
	assert.Equal(t, vec.Size(), 4)
	vec.MustAppend(5)
	assert.Equal(t, vec.Size(), 5)

	assert.Equal(t, vec.ListAll(), []int{1, 2, 3, 4, 5})

	vec.MustRemove(3)
	assert.Equal(t, vec.Size(), 4)
	assert.Equal(t, vec.ListAll(), []int{1, 2, 4, 5})

	vec.MustRemove(1)
	assert.Equal(t, vec.Size(), 3)
	assert.Equal(t, vec.ListAll(), []int{2, 4, 5})

	vec.MustRemove(2)
	assert.Equal(t, vec.Size(), 2)
	assert.Equal(t, vec.ListAll(), []int{4, 5})

	vec.MustRemove(4)
	assert.Equal(t, vec.Size(), 1)
	assert.Equal(t, vec.ListAll(), []int{5})

	vec.MustRemove(5)
	assert.Equal(t, vec.Size(), 0)
	assert.Equal(t, vec.ListAll(), []int{})
}
