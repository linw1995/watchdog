package watchdog

import (
	"context"
	"testing"
	"time"
)

type dummyDog struct{}

func (d *dummyDog) Sniff() bool {
	return true
}

func newDummyDog() *dummyDog {
	return &dummyDog{}
}

type dummySleeper struct{}

func (s *dummySleeper) Sleep() {}

func newDummySleeper() *dummySleeper {
	return &dummySleeper{}
}
func TestRun(t *testing.T) {
	sleeper := newDummySleeper()
	assertRunning := func(t *testing.T, keeper *Keeper) {
		t.Helper()
		if !keeper.Running() {
			t.Fatal("Keeper.Running() should return true")
		}
	}
	t.Run("Run success without dogs", func(t *testing.T) {
		keeper := NewKeeper()
		_, err := keeper.Run(context.Background(), sleeper)
		if err != nil {
			t.Fatal(err)
		}
		defer keeper.Cancel()
		assertRunning(t, keeper)
	})
	t.Run("Run twice without dogs", func(t *testing.T) {
		keeper := NewKeeper()
		_, err := keeper.Run(context.Background(), sleeper)
		if err != nil {
			t.Fatal(err)
		}
		defer keeper.Cancel()
		assertRunning(t, keeper)
		_, err = keeper.Run(context.Background(), sleeper)
		if err == nil || err.Error() != "&Keeper{} is already running" {
			t.Fatalf("Expect that Keep returned error but %v", err)
		}
	})
	t.Run("Run success with dogs", func(t *testing.T) {
		keeper := NewKeeper()
		err := keeper.Keep("bar", newDummyDog())
		if err != nil {
			t.Fatal(err)
		}
		_, err = keeper.Run(context.Background(), sleeper)
		if err != nil {
			t.Fatal(err)
		}
		defer keeper.Cancel()
		assertRunning(t, keeper)
	})
	t.Run("Run and collect SniffedEvent", func(t *testing.T) {
		keeper := NewKeeper()
		err := keeper.Keep("bar", newDummyDog())
		if err != nil {
			t.Fatal(err)
		}
		resultChannel, err := keeper.Run(context.Background(), sleeper)
		if err != nil {
			t.Fatal(err)
		}
		defer keeper.Cancel()
		assertRunning(t, keeper)

		select {
		case event := <-resultChannel:
			if event.dog != "bar" || !event.sniffed {
				t.Fatalf("Recieved unknown SniffedEvent %v", event)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Wait SniffedEvent recieved timeout")
		}
	})
}

func TestKeep(t *testing.T) {
	keeper := NewKeeper()
	t.Run("Keep success", func(t *testing.T) {
		err := keeper.Keep("bar", newDummyDog())
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Keep duplicated dog", func(t *testing.T) {
		err := keeper.Keep("bar", newDummyDog())
		if err == nil || err.Error() != "&Keeper{} already has Dog which name = bar" {
			t.Fatalf("Expect that Keep returned error but %v", err)
		}
	})
	t.Run("Keep after running", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := keeper.Run(ctx, newDummySleeper())
		if err != nil {
			t.Fatal(err)
		}
		err = keeper.Keep("boo", newDummyDog())
		if err == nil || err.Error() != "&Keeper{} is already running, cant keep more dogs" {
			t.Fatalf("Expect that Keep returned error but %v", err)
		}
	})
}
