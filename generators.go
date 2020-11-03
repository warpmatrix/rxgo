// Copyright 2018 The SS.SYSU Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rxgo

import (
	"context"
	"reflect"
)

// source node implementation of streamOperator
type sourceOperater struct {
	opFunc func(ctx context.Context, o *Observable, out chan interface{}) (end bool)
}

func (sop sourceOperater) op(ctx context.Context, o *Observable) {
	// must hold defintion of flow resourcs here, such as chan etc., that is allocated when connected
	// this resurces may be changed when operation routine is running.
	out := o.outflow
	//fmt.Println(o.name, "source out chan ", out)

	// Scheduler
	go func() {
		sop.opFunc(ctx, o, out)
		o.closeFlow(out)
	}()
}

func Generator(sf sourceFunc) *Observable {
	o := newGeneratorObservable("CustomSource")
	o.flip = sf
	o.operator = sourceSource
	return o
}

var sourceSource = sourceOperater{func(ctx context.Context, o *Observable, out chan interface{}) (end bool) {
	sf := o.flip.(sourceFunc)
	send := func(x interface{}) (endSignal bool) {
		endSignal = o.sendToFlow(ctx, x, out)
		return
	}
	sf(ctx, send)
	return true
}}

// creates an Observable with the provided item(s) producing by the function `func()  (val anytype, end bool)`
func Start(f interface{}) *Observable {
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{}
	outType := []reflect.Type{typeAny, typeBool}
	ctx_sup := false
	if b, cb := checkFuncUpcast(fv, inType, outType, true); !b {
		panic(ErrFuncFlip)
	} else {
		ctx_sup = cb
	}

	o := newGeneratorObservable("Start")
	o.flip_sup_ctx = ctx_sup

	o.flip = fv.Interface()
	o.operator = startSource
	return o
}

var startSource = sourceOperater{func(ctx context.Context, o *Observable, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{}
	if o.flip_sup_ctx {
		params = []reflect.Value{reflect.ValueOf(ctx)}
	}

	for end := false; !end; {
		rs, skip, stop, e := userFuncCall(fv, params)

		var item interface{}
		if stop {
			return true
		}
		if skip {
			continue
		}
		if e != nil {
			item = e
		}
		if len(rs) > 0 {
			end, _ = (rs[1].Interface()).(bool)
			item = rs[0].Interface()
		}
		// send data
		if !end {
			end = o.sendToFlow(ctx, item, out)
		}
	}

	return true
}}

// Range creates an Observable that emits a particular range of sequential integers.
func Range(start, end int) *Observable {
	o := newGeneratorObservable("Range")

	o.flip = func(ctx context.Context, out chan interface{}) {
		i := start
		for i < end {
			if b := o.sendToFlow(ctx, i, out); b {
				return
			}
			i++
		}
	}
	o.operator = rangeSource
	return o
}

var rangeSource = sourceOperater{func(ctx context.Context, o *Observable, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(out)}
	fv.Call(params)
	return true
}}

// Just creates an Observable with the provided item(s).
func Just(items ...interface{}) *Observable {
	o := newGeneratorObservable("Just")

	o.flip = func(ctx context.Context, out chan interface{}) {
		for _, item := range items {
			if b := o.sendToFlow(ctx, item, out); b {
				return
			}
		}
	}
	o.operator = justSource
	return o
}

var justSource = rangeSource
var fromSlice = rangeSource
var fromChannel = rangeSource
var fromObservable = rangeSource

// convert Slice, Channel, and Observable into Observables
func From(items interface{}) *Observable {
	v, t := reflect.ValueOf(items), reflect.TypeOf(items)

	if v.Kind() == reflect.Slice {
		length := v.Len()
		o := newGeneratorObservable("From Slice")

		o.flip = func(ctx context.Context, out chan interface{}) {
			i := 0
			for i < length {
				item := v.Index(i).Interface()
				if b := o.sendToFlow(ctx, item, out); b {
					return
				}
				i++
			}
		}
		o.operator = fromSlice
		return o
	}

	if v.Kind() == reflect.Chan {
		o := newGeneratorObservable("From Channel")

		o.flip = func(ctx context.Context, out chan interface{}) {
			for {
				// details: https://godoc.org/reflect#Select
				var selectcases = []reflect.SelectCase{
					reflect.SelectCase{Dir: reflect.SelectRecv, Chan: v},
					reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
				}
				chosen, recv, recvOK := reflect.Select(selectcases)
				if !recvOK {
					return
				}
				switch chosen {
				case 0:
				case 1:
					return
				}
				item := recv.Interface()
				if b := o.sendToFlow(ctx, item, out); b {
					return
				}
			}
		}
		o.operator = fromChannel
		return o
	}

	st := reflect.TypeOf((*Observable)(nil))
	//fmt.Println(t, st)
	if t == st {
		o := newGeneratorObservable("From *Observable")

		o.flip = func(ctx context.Context, out chan interface{}) {
			ro := v.Interface().(*Observable)
			for ; ro.next != nil; ro = ro.next {
			}
			ro.mu.Lock()
			ro.connect(ctx)
			ch := ro.outflow
			ro.mu.Unlock()
			for item := range ch {
				if b := o.sendToFlow(ctx, item, out); b {
					return
				}
			}
		}
		o.operator = fromObservable
		return o
	}

	panic(ErrFuncFlip)
}

// create an Observable that emits no items and does not terminate.
// It is important for combining with other Observables
func Never() *Observable {
	source := func(ctx context.Context, send func(x interface{}) (endSignal bool)) {
		select {
		case <-ctx.Done():
		}
	}
	o := Generator(source)
	o.Name = "Never"
	return o
}

var emptySource = rangeSource
var throwSource = rangeSource

// create an Observable that emits no items but terminates normally
func Empty() *Observable {
	o := newGeneratorObservable("Empty")

	o.flip = func(ctx context.Context, out chan interface{}) {
	}
	o.operator = emptySource
	return o
}

// create an Observable that emits no items and terminates with an error
func Throw(e error) *Observable {
	o := newGeneratorObservable("Throw")

	o.flip = func(ctx context.Context, out chan interface{}) {
		item := e
		o.sendToFlow(ctx, item, out)
	}
	o.operator = throwSource
	return o
}

func newGeneratorObservable(name string) (o *Observable) {
	//new Observable
	o = newObservable()
	o.Name = name

	//chain Observables
	o.root = o

	//set options
	o.buf_len = 0
	return o
}
