package main

import (
	"os/exec"
	"testing"
)

func TestProcessAliveDog(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	cmd.Start()
	p := cmd.Process

	dog := NewProcessAliveDog(p.Pid)
	want := true
	got := dog.Sniff()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	p.Kill()
	p.Wait()

	want = false
	got = dog.Sniff()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
