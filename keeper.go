package watchdog

import (
	"context"
	"fmt"
)

type DogHandle struct {
	dog     Dog
	channel chan bool
	Cancel  context.CancelFunc
}

type Keeper struct {
	name2DogHandle map[string]*DogHandle

	ctx    context.Context
	Cancel context.CancelFunc
}

type SniffedEvent struct {
	dog     string
	sniffed bool
}

func (k *Keeper) String() string {
	return "&Keeper{}"
}

// Flag repressents the Keeper is running.
func (k *Keeper) Running() bool {
	return k.ctx != nil
}

// Keeper keeps dogs for sniffing.
func (k *Keeper) Keep(name string, dog Dog) error {
	if k.Running() {
		// TODO: allow to add more dogs
		return fmt.Errorf("%v is already running, cant keep more dogs", k)
	}

	_, ok := k.name2DogHandle[name]
	if ok {
		return fmt.Errorf("%v already has Dog which name = %v", k, name)
	}
	k.name2DogHandle[name] = &DogHandle{dog: dog, channel: make(chan bool)}
	return nil
}

// Keeper runs and collects SniffedEvent from dogs sniffing.
func (k *Keeper) Run(ctx context.Context, sleeper Sleeper) (<-chan SniffedEvent, error) {
	if k.Running() {
		return nil, fmt.Errorf("%v is already running", k)
	}
	k.ctx, k.Cancel = context.WithCancel(ctx)

	resultChannel := make(chan SniffedEvent)
	collectingSniffedEvent := func(ctx context.Context, result chan SniffedEvent, target chan bool, name string) {
		for {
			select {
			case sniffed := <-target:
				result <- SniffedEvent{dog: name, sniffed: sniffed}
			case <-ctx.Done():
				return
			}
		}
	}

	for name, handle := range k.name2DogHandle {
		handleCtx, cancel := context.WithCancel(ctx)
		handle.Cancel = cancel
		go collectingSniffedEvent(handleCtx, resultChannel, handle.channel, name)
		go Sniffing(handleCtx, handle.dog, sleeper, handle.channel)
	}
	return resultChannel, nil
}

func NewKeeper() *Keeper {
	return &Keeper{name2DogHandle: make(map[string]*DogHandle)}
}
