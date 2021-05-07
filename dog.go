package main

import (
	"context"
	"os"
	"syscall"
)

type Sleeper interface {
	Sleep()
}

type Dog interface {
	Sniff() bool
}

// Keep sniffing
func Sniffing(ctx context.Context, dog Dog, sleeper Sleeper, resultChannel chan bool) {
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

// Create a ProcessAliveDog via the Pid of a process it is watching.
func NewProcessAliveDog(pid int) *ProcessAliveDog {
	return &ProcessAliveDog{Pid: pid}
}
