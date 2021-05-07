package main

import (
	"context"
	"os/exec"
	"testing"
)

type dummyProcess struct {
	cmd *exec.Cmd
}

func (dp *dummyProcess) Start() {
	err := dp.cmd.Start()
	if err != nil {
		panic(err)
	}
}

func (dp *dummyProcess) Stop() {
	err := dp.cmd.Process.Kill()
	if err != nil {
		panic(err)
	}
	_, err = dp.cmd.Process.Wait()
	if err != nil {
		panic(err)
	}
}

func (dp *dummyProcess) Pid() int {
	return dp.cmd.Process.Pid
}

func newDummyProcess(name string, args ...string) *dummyProcess {
	return &dummyProcess{cmd: exec.Command(name, args...)}
}

type sleeperWithControl struct {
	controlChannel chan interface{}
}

func (s *sleeperWithControl) Sleep() {
	<-s.controlChannel
}

func (s *sleeperWithControl) Awake() {
	s.controlChannel <- nil
}

func newSleeperWithControl() *sleeperWithControl {
	return &sleeperWithControl{controlChannel: make(chan interface{})}
}

type dogWithControl struct {
	flag bool
}

func (dog *dogWithControl) Sniff() bool {
	return dog.flag
}

func (dog *dogWithControl) Sniffed() {
	dog.flag = true
}

func (dog *dogWithControl) Unsniffed() {
	dog.flag = false
}

func newDogWithControl(initFlag bool) *dogWithControl {
	return &dogWithControl{flag: initFlag}
}

func TestSniffing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dog := newDogWithControl(true)
	sleeper := newSleeperWithControl()
	resultChannel := make(chan bool)

	go Sniffing(ctx, dog, sleeper, resultChannel)

	want := true
	got := <-resultChannel
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	dog.Unsniffed()
	sleeper.Awake()
	want = false
	got = <-resultChannel
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestProcessAliveDog(t *testing.T) {
	t.Run("Check if process is alive or not via its Pid", func(t *testing.T) {
		dp := newDummyProcess("sleep", "60")
		dp.Start()

		dog := NewProcessAliveDog(dp.Pid())
		want := true
		got := dog.Sniff()
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		dp.Stop()

		want = false
		got = dog.Sniff()
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

	})

	t.Run("Keep checking", func(t *testing.T) {
		dp := newDummyProcess("sleep", "60")
		dp.Start()
		dog := NewProcessAliveDog(dp.Pid())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		isProcessAliveChannel := make(chan bool)

		sleeper := newSleeperWithControl()
		go Sniffing(ctx, dog, sleeper, isProcessAliveChannel)

		want := true
		got := <-isProcessAliveChannel
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		dp.Stop()
		sleeper.Awake()

		want = false
		got = <-isProcessAliveChannel
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

	})
}
