package rxgo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/warpmatrix/rxgo"
)

type observer struct {
	name string
}

func (o observer) OnNext(x interface{}) {
	fmt.Println(o.name, "observed value ", x)
}

func (o observer) OnError(e error) {
	fmt.Println(o.name, "Error ", e)
}

func (o observer) OnCompleted() {
	fmt.Println(o.name, "Down ")
}

func TestMain(t *testing.T) {

	// test Subscribe on any
	ob := rxgo.Just(10, 20, 30).Map(dd)
	ob1 := ob.Map(dd).SubscribeOn(rxgo.ThreadingIO).Debug(true).Map(dd)
	ob1.Subscribe(func(x int) {
		fmt.Println("Just", x)
	})

	ob = rxgo.Just(0, 12, 7, 34, 2).Filter(func(x int) bool {
		return x < 10
	}).SubscribeOn(rxgo.ThreadingIO)
	ob.Subscribe(
		func(x int) {
			fmt.Println("Filter", x)
		})
}

func dd(x int) int { return 2 * x }

func TestObserver(t *testing.T) {
	var s rxgo.Observer = observer{"test observer"}
	rxgo.Just(1, 2, 3).Subscribe(s)
}

func TestTreading(t *testing.T) {
	flow := rxgo.Just(10, 20, 30).Map(func(x int) int {
		return x + 1
	})
	/* 	.FlatMap(func(x int) *rxgo.Observable {
		return rxgo.Just(x+1, x+2)
	}).SubscribeOn(rxgo.ThreadingIO) */

	go flow.Subscribe(observer{"test flatMap"})
	//time.Sleep(time.Nanosecond * 1000)
	flow.Subscribe(observer{"test flatMap again"})
	time.Sleep(time.Microsecond * 1000)
}
