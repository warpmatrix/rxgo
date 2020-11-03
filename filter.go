package rxgo

import (
	"context"
	"reflect"
)

// filter node implementation of streamOperator
type filtOperater struct {
	opFunc func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool)
}

func (fop filtOperater) op(ctx context.Context, o *Observable) {
	in := o.pred.outflow
	out := o.outflow

	// Scheduler
	go func() {
		for end := false; !end; { // made panic op re-enter
			end = fop.opFunc(ctx, o, in, out)
		}
		o.closeFlow(out)
	}()
}

// First emit only the first item, or the first item that meets a condition, from an Observable
func (parent *Observable) First(f interface{}) (o *Observable) {
	o = parent.newFilterObservable("First")
	fv := reflect.ValueOf(f)
	inType := []reflect.Type{typeAny}
	outType := []reflect.Type{typeBool}
	b, ctxSup := checkFuncUpcast(fv, inType, outType, true)
	if !b {
		panic(ErrFuncFlip)
	}
	o.flip_accept_error = checkFuncAcceptError(fv)
	o.flip_sup_ctx = ctxSup
	o.flip = fv.Interface()
	o.operator = firstOperator
	return o
}

var firstOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	for x := range in {
		if end {
			break
		}
		xv := reflect.ValueOf(x)
		params := []reflect.Value{xv}
		rs, skip, stop, e := userFuncCall(fv, params)
		var item interface{} = rs[0].Interface()
		if stop {
			end = true
			continue
		}
		if skip {
			continue
		}
		if e != nil {
			item = e
		}
		// send data
		if !end {
			if b, ok := item.(bool); ok && b {
				o.sendToFlow(ctx, xv.Interface(), out)
				end = true
			}
		}
	}
	return
}}

func (parent *Observable) newFilterObservable(name string) (o *Observable) {
	//new Observable
	o = newObservable()
	o.Name = name

	//chain Observables
	parent.next = o
	o.pred = parent
	o.root = parent.root

	//set options
	// o.buf_len = BufferLen
	return o
}
