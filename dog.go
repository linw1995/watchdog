package main

import (
	"context"
	"os"
	"syscall"
)

type Sleeper interface {
	Sleep()
}

// Dog for checking if process is alive or not via its Pid.
type ProcessAliveDog struct {
	Pid int
}

// Check if process is alive or not via its Pid.
func (dog ProcessAliveDog) Sniff() bool {
	p, _ := os.FindProcess(dog.Pid)
	err := p.Signal(syscall.Signal(0))
	return err == nil
}

// Keep checking if process is alive or not via its Pid.
func (dog ProcessAliveDog) Sniffing(ctx context.Context, sleeper Sleeper, resultChannel chan bool) {
	for {
		select {
		case <-ctx.Done():
			break
		default:
			resultChannel <- dog.Sniff()
		}
		sleeper.Sleep()
	}
}

// Create a ProcessAliveDog via the Pid of a process it is watching.
func NewProcessAliveDog(pid int) *ProcessAliveDog {
	return &ProcessAliveDog{Pid: pid}
}
