package demux

import "sync"

// Dynamic creates dynamic demultiplexer that routes items from 'in' based on keys returned by 'keyFunc'.
// For each unique key, a new goroutine is spawned running 'consumeFunc'.
// Each consumeFunc receives a channel that delivers values matching its key.
func Dynamic[T any, K comparable](
	in <-chan T,
	keyFunc func(T) K,
	consumeFunc func(K, <-chan T),
) {
	outChans := make(map[K]chan T)
	var wg sync.WaitGroup

	for t := range in {
		key := keyFunc(t)
		ch, exists := outChans[key]
		if !exists {
			ch = make(chan T)
			outChans[key] = ch

			wg.Add(1)
			go func() {
				defer wg.Done()
				consumeFunc(key, ch)
			}()
		}
		ch <- t
	}

	for _, ch := range outChans {
		close(ch)
	}

	wg.Wait()
}
