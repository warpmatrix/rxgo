package rxgo_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/warpmatrix/rxgo"
)

func TestDebounce(t *testing.T) {
	res := []int{}
	timespan := 2 * time.Millisecond
	rxgo.Just(0, 1, 2, 3, 4, 5).Map(func(x int) int {
		time.Sleep(1 * time.Millisecond)
		return x
	}).Debounce(timespan).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{0, 2, 4}, res, "Debounce Test Error!")
}

func TestDistinct(t *testing.T) {
	res := []int{}
	rxgo.Just(1, 2, 2, 1, 3).Distinct(func(item interface{}) interface{} {
		return item
	}).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{1, 2, 3}, res, "Distinct Test Error!")
}

func TestElementAt(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).ElementAt(2).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{2}, res, "ElementAt Test Error!")
}

func TestIgnoreElements(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).IgnoreElements().Subscribe(func(x int) {
		res = append(res, x)
	})
	assert.Equal(t, []int{}, res, "IgnoreElements Test Error!")
}

func TestFirst(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).First(func(x int) bool {
		return x > 2
	}).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{3}, res, "First Test Error!")
}

func TestLast(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).Last(
		func(x int) bool { return true },
	).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{6}, res, "Last Test Error!")
}

func TestSample(t *testing.T) {
	res := []int{}
	smpChan := make(chan interface{})
	go func() {
		rxgo.Just(4, 3, 1, 5, 1).Map(func(x int) int {
			delay := time.Duration(x) * time.Millisecond
			time.Sleep(delay)
			return x
		}).Subscribe(func(x int) {
			smpChan <- x
		})
	}()

	rxgo.Just(1, 2, 3, 4, 5).Map(func(x int) int {
		time.Sleep(3 * time.Millisecond)
		return x
	}).Sample(smpChan).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{1, 2, 4}, res, "Sample Test Error!")
}

func TestSkip(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4).Skip(3).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{3, 4}, res, "Skip Test Error!")
}

func TestSkipLast(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4).SkipLast(3).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{0, 1}, res, "SkipLast Test Error!")
}

func TestTake(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).Take(4).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{0, 1, 2, 3}, res, "Take Test Error!")
}

func TestTakeLasr(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 1, 2, 3, 4, 5, 6).TakeLast(4).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{3, 4, 5, 6}, res, "TakeLast Test Error!")
}
