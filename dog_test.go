package main

import (
	"context"
	"os/exec"
	"testing"
	"time"
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

func (s *sleeperWithControl) AwakeOnce() {
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
	sleeper.AwakeOnce()
	want = false
	got = <-resultChannel
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWaitSniffing(t *testing.T) {
	tryReceive := func(channel chan interface{}, wait time.Duration) bool {
		if wait > 0 {
			select {
			case <-channel:
				return true
			case <-time.After(wait):
				return false
			}
		} else {
			select {
			case <-channel:
				return true
			default:
				return false
			}
		}
	}
	type SubTestParam struct {
		Description string
		InitFlag    bool
		Target      func(context.Context, Dog, Sleeper) chan interface{}
	}
	subTestParams := [...]SubTestParam{
		{"Test WaitSniffed", false, WaitSniffed},
		{"Test WaitUnsniffed", true, WaitUnSniffed},
	}
	for _, param := range subTestParams {
		t.Run(param.Description, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dog := newDogWithControl(param.InitFlag)
			sleeper := newSleeperWithControl()

			doneChannel := param.Target(ctx, dog, sleeper)
			want := false
			got := tryReceive(doneChannel, 0)
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			for i := 0; i < 10; i++ {
				sleeper.AwakeOnce()
				got = tryReceive(doneChannel, 0)
				if got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			}

			if param.InitFlag {
				dog.Unsniffed()
			} else {
				dog.Sniffed()
			}
			sleeper.AwakeOnce()
			want = true
			got = tryReceive(doneChannel, 1*time.Second)
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}

		})
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
		sleeper.AwakeOnce()

		want = false
		got = <-isProcessAliveChannel
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

	})
}

func BenchmarkProcessAliveDogSniff(b *testing.B) {
	dp := newDummyProcess("sleep", "3600")
	dp.Start()
	dog := NewProcessAliveDog(dp.Pid())
	for i := 0; i < b.N; i++ {
		dog.Sniff()
	}
	dp.Stop()
}
