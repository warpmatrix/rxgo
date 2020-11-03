package rxgo

import (
	"context"
	"reflect"
	"time"
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
		fop.opFunc(ctx, o, in, out)
		o.closeFlow(out)
	}()
}

// Debounce only emit an item from an Observable if a particular timespan has passed without it emitting another item
func (parent *Observable) Debounce(timespan time.Duration) (o *Observable) {
	o = parent.newFilterObservable("debounce")
	o.flip = func(ctx context.Context, in chan interface{}, out chan interface{}) (end bool) {
		var latest reflect.Value
		go func() {
			for !end {
				than := time.After(timespan)
				<-than
				if latest != reflect.ValueOf(nil) {
					if o.sendToFlow(ctx, latest.Interface(), out) {
						end = true
					}
					latest = reflect.ValueOf(nil)
				}
			}
			o.closeFlow(out)
		}()

		for !end {
			select {
			case <-ctx.Done():
				end = true
			case item, ok := <-in:
				if !ok {
					end = true
				}
				latest = reflect.ValueOf(item)
			}
		}
		return
	}
	o.operator = debounceOperator
	return o
}

var debounceOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in), reflect.ValueOf(out)}
	fv.Call(params)
	return true
}}

// used in distinct operator
type cmpKeyFunc func(interface{}) interface{}

// Distinct suppress duplicate items emitted by an Observable
func (parent *Observable) Distinct(f cmpKeyFunc) (o *Observable) {
	o = parent.newFilterObservable("distinct")
	o.flip = f
	o.operator = distinctOperator
	return o
}

var distinctOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	keyset := make(map[interface{}]struct{})
	fv := o.flip.(cmpKeyFunc)
	for !end {
		select {
		case <-ctx.Done():
			end = true
		case item, ok := <-in:
			if !ok {
				end = true
				break
			}
			latest := reflect.ValueOf(item)
			key := fv(latest.Interface())
			_, hasKey := keyset[key]
			if !hasKey {
				keyset[key] = struct{}{}
				if o.sendToFlow(ctx, latest.Interface(), out) {
					return
				}
			}
		}
	}
	return
}}

// ElementAt emit only item n emitted by an Observable
func (parent *Observable) ElementAt(n uint) (o *Observable) {
	o = parent.newFilterObservable("ElementAt")
	o.flip = func(ctx context.Context, in chan interface{}, out chan interface{}) (end bool) {
		i := uint(0)
		for !end {
			select {
			case <-ctx.Done():
				end = true
			case item, ok := <-in:
				if !ok {
					end = true
					break
				}
				latest := reflect.ValueOf(item)
				if i == n {
					if o.sendToFlow(ctx, latest.Interface(), out) {
						end = true
						return
					}
				}
				i++
			}
		}
		return
	}
	o.operator = elementAtOperator
	return o
}

var elementAtOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	params := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in), reflect.ValueOf(out)}
	fv.Call(params)
	return true
}}

// IgnoreElements do not emit any items from an Observable but mirror its termination notification
func (parent *Observable) IgnoreElements() (o *Observable) {
	o = parent.newFilterObservable("IgnoreElements")
	o.operator = ignoreElementsOperator
	return o
}

var ignoreElementsOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	return true
}}

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

// Last emit only the last item emitted by an Observable
func (parent *Observable) Last(f interface{}) (o *Observable) {
	o = parent.newFilterObservable("Last")
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
	o.operator = lastOperator
	return o
}

var lastOperator = filtOperater{func(ctx context.Context, o *Observable, in chan interface{}, out chan interface{}) (end bool) {
	fv := reflect.ValueOf(o.flip)
	var last reflect.Value
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
				last = xv
			}
		}
	}
	if last != reflect.ValueOf(nil) {
		end = o.sendToFlow(ctx, last.Interface(), out)
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
	return o
}
