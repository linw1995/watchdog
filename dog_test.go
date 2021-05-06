package main

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestProcessAliveDog(t *testing.T) {
	t.Run("Check if process is alive or not via its Pid",
		func(t *testing.T) {
			cmd := exec.Command("sleep", "60")
			err := cmd.Start()
			if err != nil {
				t.Errorf(`Run command "sleep 60" error %v`, err)
			}

			p := cmd.Process
			dog := NewProcessAliveDog(p.Pid)
			want := true
			got := dog.Sniff()
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			err = p.Kill()
			if err != nil {
				t.Errorf(`Kill command "sleep 60" error %v`, err)
			}
			_, err = p.Wait()
			if err != nil {
				t.Errorf(`Wait command "sleep 60" error %v`, err)
			}

			want = false
			got = dog.Sniff()
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}

		},
	)

	t.Run("Keep checking", func(t *testing.T) {
		cmd := exec.Command("sleep", "60")
		err := cmd.Start()
		if err != nil {
			t.Errorf(`Run command "sleep 60" error %v`, err)
		}

		p := cmd.Process
		dog := NewProcessAliveDog(p.Pid)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		isProcessAliveChannel := make(chan bool)

		go dog.Sniffing(ctx, 1 * time.Second, isProcessAliveChannel)

		want := true
		got := <-isProcessAliveChannel
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		err = p.Kill()
		if err != nil {
			t.Errorf(`Kill command "sleep 60" error %v`, err)
		}
		_, err = p.Wait()
		if err != nil {
			t.Errorf(`Wait command "sleep 60" error %v`, err)
		}

		want = false
		got = <-isProcessAliveChannel
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

	})
}
