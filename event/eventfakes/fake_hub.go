// This file was generated by counterfeiter
package eventfakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
)

type FakeHub struct {
	EmitStub        func(receptor.Event)
	emitMutex       sync.RWMutex
	emitArgsForCall []struct {
		arg1 receptor.Event
	}
	SubscribeStub        func() receptor.EventSource
	subscribeMutex       sync.RWMutex
	subscribeArgsForCall []struct{}
	subscribeReturns     struct {
		result1 receptor.EventSource
	}
}

func (fake *FakeHub) Emit(arg1 receptor.Event) {
	fake.emitMutex.Lock()
	fake.emitArgsForCall = append(fake.emitArgsForCall, struct {
		arg1 receptor.Event
	}{arg1})
	fake.emitMutex.Unlock()
	if fake.EmitStub != nil {
		fake.EmitStub(arg1)
	}
}

func (fake *FakeHub) EmitCallCount() int {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return len(fake.emitArgsForCall)
}

func (fake *FakeHub) EmitArgsForCall(i int) receptor.Event {
	fake.emitMutex.RLock()
	defer fake.emitMutex.RUnlock()
	return fake.emitArgsForCall[i].arg1
}

func (fake *FakeHub) Subscribe() receptor.EventSource {
	fake.subscribeMutex.Lock()
	fake.subscribeArgsForCall = append(fake.subscribeArgsForCall, struct{}{})
	fake.subscribeMutex.Unlock()
	if fake.SubscribeStub != nil {
		return fake.SubscribeStub()
	} else {
		return fake.subscribeReturns.result1
	}
}

func (fake *FakeHub) SubscribeCallCount() int {
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	return len(fake.subscribeArgsForCall)
}

func (fake *FakeHub) SubscribeReturns(result1 receptor.EventSource) {
	fake.SubscribeStub = nil
	fake.subscribeReturns = struct {
		result1 receptor.EventSource
	}{result1}
}

var _ event.Hub = new(FakeHub)