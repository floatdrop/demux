package demux_test

import (
	"sync"
	"testing"
	"time"

	"github.com/floatdrop/demux"
)

func TestDynamic_BasicDemuxing(t *testing.T) {
	type input struct {
		id   int
		data string
	}

	in := make(chan input)
	var mu sync.Mutex
	result := make(map[int][]string)

	go demux.Dynamic(in,
		func(i input) int { return i.id },
		func(k int, ch <-chan input) {
			var items []string
			for item := range ch {
				items = append(items, item.data)
			}
			mu.Lock()
			result[k] = items
			mu.Unlock()
		})

	// Send data
	go func() {
		in <- input{id: 1, data: "a"}
		in <- input{id: 2, data: "b"}
		in <- input{id: 1, data: "c"}
		in <- input{id: 2, data: "d"}
		close(in)
	}()

	time.Sleep(100 * time.Millisecond)

	// Assertions
	if len(result) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result))
	}
	if len(result[1]) != 2 || result[1][0] != "a" || result[1][1] != "c" {
		t.Errorf("unexpected data for key 1: %v", result[1])
	}
	if len(result[2]) != 2 || result[2][0] != "b" || result[2][1] != "d" {
		t.Errorf("unexpected data for key 2: %v", result[2])
	}
}

func TestDynamic_SingleKey(t *testing.T) {
	in := make(chan int)
	var collected []int
	done := make(chan struct{})

	go demux.Dynamic(
		in,
		func(i int) string { return "only" },
		func(k string, ch <-chan int) {
			for i := range ch {
				collected = append(collected, i)
			}
			close(done)
		})

	go func() {
		in <- 10
		in <- 20
		in <- 30
		close(in)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for consumer")
	}

	expected := []int{10, 20, 30}
	for i, val := range expected {
		if collected[i] != val {
			t.Errorf("expected %d at index %d, got %d", val, i, collected[i])
		}
	}
}

func TestDynamic_NoInput(t *testing.T) {
	in := make(chan string)
	called := false

	go demux.Dynamic(
		in,
		func(s string) int { return 0 },
		func(k int, ch <-chan string) {
			called = true
		},
	)

	close(in)

	time.Sleep(100 * time.Millisecond)
	if called {
		t.Error("consumeFunc should not be called for empty input")
	}
}
