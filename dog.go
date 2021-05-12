package watchdog

import (
	"context"
	"os"
	"syscall"
)

// Sleeper for controling the delay duration.
type Sleeper interface {
	Sleep()
}

// Dog for monitoring specific status changes.
type Dog interface {
	Sniff() bool
}

// Sniffing forever.
func Sniffing(ctx context.Context, dog Dog, sleeper Sleeper, resultChannel chan bool) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resultChannel <- dog.Sniff()
		}
		sleeper.Sleep()
	}
}

// WaitSniffResult keeps sniffing until expected result received.
func WaitSniffResult(ctx context.Context, dog Dog, sleeper Sleeper, result bool) chan interface{} {
	resultChannel := make(chan interface{})
	sniffResultChannel := make(chan bool)
	sniffingCtx, cancel := context.WithCancel(ctx)
	go func(sniffResultChannel chan bool,
		resultChannel chan interface{},
		sniffingCtx context.Context,
		cancel context.CancelFunc,
	) {
		defer cancel()
		for {
			select {
			case sniffed := <-sniffResultChannel:
				if sniffed == result {
					resultChannel <- nil
					return
				}
			case <-sniffingCtx.Done():
				close(resultChannel)
				return
			}
		}
	}(sniffResultChannel, resultChannel, sniffingCtx, cancel)
	go Sniffing(sniffingCtx, dog, sleeper, sniffResultChannel)
	return resultChannel

}

// WaitSniffed keeps sniffing until sniffed.
func WaitSniffed(ctx context.Context, dog Dog, sleeper Sleeper) chan interface{} {
	return WaitSniffResult(ctx, dog, sleeper, true)
}

// WaitUnSniffed keeps sniffing until unsinffed.
func WaitUnSniffed(ctx context.Context, dog Dog, sleeper Sleeper) chan interface{} {
	return WaitSniffResult(ctx, dog, sleeper, false)
}

// ProcessAliveDog for checking if process is alive or not via its Pid.
type ProcessAliveDog struct {
	Pid int
}

// Sniff checks if process is alive or not via its Pid.
func (dog ProcessAliveDog) Sniff() bool {
	p, _ := os.FindProcess(dog.Pid)
	err := p.Signal(syscall.Signal(0))
	return err == nil
}

// NewProcessAliveDog creates a ProcessAliveDog via the Pid of a process it is watching.
func NewProcessAliveDog(pid int) *ProcessAliveDog {
	return &ProcessAliveDog{Pid: pid}
}
