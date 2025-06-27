package demux_test

import (
	"sync"
	"testing"
	"time"

	"github.com/floatdrop/demux"
)

func TestStatic_RoutesToCorrectChannels(t *testing.T) {
	in := make(chan int)
	outA := make(chan int, 2)
	outB := make(chan int, 2)

	channels := map[string]chan<- int{
		"A": outA,
		"B": outB,
	}

	keyFunc := func(i int) string {
		if i%2 == 0 {
			return "A"
		}
		return "B"
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		demux.Static(in, keyFunc, channels)
	}()

	// Send data
	go func() {
		in <- 1
		in <- 2
		in <- 3
		in <- 4
		close(in)
	}()

	wg.Wait()

	// Collect output
	var aValues, bValues []int

	done := make(chan struct{})
	go func() {
		for v := range outA {
			aValues = append(aValues, v)
		}
		for v := range outB {
			bValues = append(bValues, v)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}

	// Validate
	expectedA := []int{2, 4}
	expectedB := []int{1, 3}

	if !equal(aValues, expectedA) {
		t.Errorf("expected A channel values %v, got %v", expectedA, aValues)
	}
	if !equal(bValues, expectedB) {
		t.Errorf("expected B channel values %v, got %v", expectedB, bValues)
	}
}

// Helper for slice comparison
func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
