package demux

import (
	"container/list"
	"math"
	"sync"
)

type channelInfo[T any, K comparable] struct {
	ch      chan T
	key     K
	element *list.Element
}

// Config holds configuration for the dynamic demultiplexer
type Config struct {
	maxChannels int
	bufferSize  int
}

// Option is a functional option for configuring the demultiplexer
type Option func(*Config)

// WithMaxChannels sets the maximum number of concurrent channels
func WithMaxChannels(max int) Option {
	return func(c *Config) {
		if max > 0 {
			c.maxChannels = max
		}
	}
}

// WithBufferSize sets the buffer size for each channel
func WithBufferSize(size int) Option {
	return func(c *Config) {
		if size >= 0 {
			c.bufferSize = size
		}
	}
}

// defaultConfig returns the default configuration
func defaultConfig() Config {
	return Config{
		maxChannels: math.MaxInt, // uncapped by default
		bufferSize:  0,           // unbuffered by default
	}
}

// Dynamic creates dynamic demultiplexer that routes items from 'in' based on keys returned by 'keyFunc'.
// For each unique key, a new goroutine is spawned running 'consumeFunc'.
// Each consumeFunc receives a channel that delivers values matching its key.
// When maxChannels limit is reached, least recently used channels are evicted.
func Dynamic[T any, K comparable](
	in <-chan T,
	keyFunc func(T) K,
	consumeFunc func(K, <-chan T),
	opts ...Option,
) {
	// Apply options
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	outChans := make(map[K]*channelInfo[T, K])
	lru := list.New()
	var wg sync.WaitGroup

	for t := range in {
		key := keyFunc(t)

		info, exists := outChans[key]

		if exists {
			// Move to front (most recently used)
			lru.MoveToFront(info.element)
		} else {
			// Check if we need to evict
			if len(outChans) >= config.maxChannels {
				// Evict least recently used
				oldest := lru.Back()
				if oldest != nil {
					oldInfo := oldest.Value.(*channelInfo[T, K])
					close(oldInfo.ch)
					delete(outChans, oldInfo.key)
					lru.Remove(oldest)
				}
			}

			// Create new channel with configured buffer size
			ch := make(chan T, config.bufferSize)
			element := lru.PushFront(nil)
			info = &channelInfo[T, K]{
				ch:      ch,
				key:     key,
				element: element,
			}
			element.Value = info
			outChans[key] = info

			wg.Add(1)
			go func(k K, c <-chan T) {
				defer wg.Done()
				consumeFunc(k, c)
			}(key, ch)
		}

		info.ch <- t
	}

	// Close all remaining channels
	for _, info := range outChans {
		close(info.ch)
	}

	wg.Wait()
}
