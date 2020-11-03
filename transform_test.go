package rxgo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warpmatrix/rxgo"
)

func TestAnyTranform(t *testing.T) {
	iCount := 0
	sCount := 0
	eCount := 0
	res := []int{}

	rxgo.Generator(func(ctx context.Context, send func(x interface{}) (endSignal bool)) {
		send(10)
		send("hello")
		send(20)
		send(errors.New("Any"))
		send(30)
	}).TransformOp(func(ctx context.Context, item interface{}, send func(x interface{}) (endSignal bool)) {
		if i, ok := item.(int); ok {
			send(i + 1)
		} else {
			send(item)
		}
	}).Subscribe(rxgo.ObserverMonitor{
		Next: func(item interface{}) {
			if i, ok := item.(int); ok {
				iCount++
				res = append(res, i)
			} else {
				sCount++
			}
		},
		Error: func(e error) {
			eCount++
		},
	})

	assert.Equal(t, []int{3, 1, 1}, []int{iCount, sCount, eCount}, "type count error")
	assert.Equal(t, []int{11, 21, 31}, res, "transform data error")
}

func TestMap(t *testing.T) {
	res := []int{}
	ob := rxgo.Just(10, 20, 30).Map(func(x int) int {
		return 2 * x
	})
	ob.Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{20, 40, 60}, res, "Map Test Error!")

	res1 := []interface{}{}
	ee := errors.New("Any")
	rxgo.Generator(func(ctx context.Context, send func(x interface{}) (endSignal bool)) {
		send(10)
		send(ee)
		send(30)
	}).Map(func(x int) int {
		return 2 * x
	}).Subscribe(rxgo.ObserverMonitor{
		Next: func(item interface{}) {
			res1 = append(res1, item)
		},
		Error: func(e error) {
			res1 = append(res1, e)
		},
	})

	assert.Equal(t, []interface{}{20, ee, 60}, res1, "Map1 Test Error!")
}

func TestFlatMap(t *testing.T) {
	res := []int{}
	rxgo.Just(10, 20, 30).FlatMap(func(x int) *rxgo.Observable {
		return rxgo.Just(x+1, x+2)
	}).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{11, 12, 21, 22, 31, 32}, res, "Map Test Error!")
}

func TestFilter(t *testing.T) {
	res := []int{}
	rxgo.Just(0, 12, 7, 34, 2).Filter(func(x int) bool {
		return x < 10
	}).Subscribe(func(x int) {
		res = append(res, x)
	})

	assert.Equal(t, []int{0, 7, 2}, res, "Filter Test Error!")
}
