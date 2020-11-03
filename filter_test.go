package rxgo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warpmatrix/rxgo"
)

func TestObservableFirst(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).First(
		func(x int) bool {
			return x > 2
		}).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{3}, res, "Filter Test Error!")
}
