package utils

import "sync/atomic"

type OnceValue struct {
	value atomic.Value
	done  chan struct{}
}

func NewOnceValue() *OnceValue {
	ov := &OnceValue{
		done: make(chan struct{}, 1),
	}
	return ov
}

func (ov *OnceValue) Set(value interface{}) {
	ov.value.Store(value)
	close(ov.done)
}

func (ov *OnceValue) Get() interface{} {
	<-ov.done
	return ov.value.Load()
}
