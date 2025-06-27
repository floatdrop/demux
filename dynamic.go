package demux

// Dynamic creates dynamic demultiplexer that routes items from 'in' based on keys returned by 'keyFunc'.
// For each unique key, a new goroutine is spawned running 'consumeFunc'.
// Each consumeFunc receives a channel that delivers values matching its key.
func Dynamic[T any, K comparable](
	in <-chan T,
	keyFunc func(T) K,
	consumeFunc func(K, <-chan T),
) {
	outChans := make(map[K]chan T)
	defer func() {
		for _, ch := range outChans {
			close(ch)
		}
	}()

	for t := range in {
		key := keyFunc(t)
		ch, exists := outChans[key]
		if !exists {
			ch = make(chan T)
			outChans[key] = ch

			go consumeFunc(key, ch)
		}
		ch <- t
	}
}
