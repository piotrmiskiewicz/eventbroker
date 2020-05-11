package main

import (
	"context"
	"reflect"
	"fmt"
	"strings"
)

type EventHandler = func(ctx context.Context, ev interface{}) error

type Publisher interface {
	Publish(ctx context.Context, event interface{})
}

type ApplicationEventBroker struct {
	handlers map[reflect.Type][]EventHandler
}

func (b *ApplicationEventBroker) Publish(ctx context.Context, ev interface{}) {
	tt := reflect.TypeOf(ev)
	hList, found := b.handlers[tt]
	if found {
		for _, h := range hList {
			h(ctx, ev)
		}
	}
}

func (b *ApplicationEventBroker) Subscribe(evType interface{}, evHandler EventHandler) {

	tt := reflect.TypeOf(evType)
	// todo: sync

	if _, found := b.handlers[tt]; !found {
		b.handlers[tt] = []EventHandler{}
	}

	b.handlers[tt] = append(b.handlers[tt], evHandler)
}

type (
	ProvisioningOperation struct {
		Name string
	}

	Started struct {
		op ProvisioningOperation
	}

	Failed struct {
		op ProvisioningOperation
	}

	Succeeded struct {
		op ProvisioningOperation
	}
)

type SomeBusinessLogic struct {
	publisher Publisher
}

func (s *SomeBusinessLogic) Execute(n string) {
	fmt.Println("Processing ", n)
	obj := ProvisioningOperation{Name: n}

	s.publisher.Publish(context.TODO(), Started{obj})

	if strings.HasPrefix(n, "F") {
		fmt.Println("Failing")
		s.publisher.Publish(context.TODO(), Failed{obj})
		return
	}

	s.publisher.Publish(context.TODO(), Succeeded{obj})
}

type MetricsAggregator struct {
	// todo: sync
	Succeeded int
	Failed    int
}

func (a *MetricsAggregator) OnFailed(ctx context.Context, ev interface{}) error {
	// if we really need the event object - cast here
	//todo: sync
	a.Failed = a.Failed + 1
	return nil
}

func (a *MetricsAggregator) OnSucceeded(ctx context.Context, ev interface{}) error {
	//todo: sync
	a.Succeeded = a.Succeeded + 1
	return nil
}

func main() {
	ma := &MetricsAggregator{}

	aeb := &ApplicationEventBroker{handlers: make(map[reflect.Type][]EventHandler)}
	aeb.Subscribe(Failed{}, ma.OnFailed)
	aeb.Subscribe(Succeeded{}, ma.OnSucceeded)

	svc := SomeBusinessLogic{publisher: aeb}

	svc.Execute("One")   // ok
	svc.Execute("Two")   // ok
	svc.Execute("F one") // will fail

	fmt.Println("Succedeed: ", ma.Succeeded)
	fmt.Println("Failed: ", ma.Failed)

}
