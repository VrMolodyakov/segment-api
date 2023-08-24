package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAnySliceConversion(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	arrayAsAny := ToAnySlice[int](array)
	for i := range array {
		got, ok := arrayAsAny[i].(int)
		assert.True(t, ok)
		assert.Equal(t, array[i], got)
	}
}
