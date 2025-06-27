package demux

// Static creates Static demultipler that routes each element to channel in channels map by key
// computed by keyFunc.
func Static[T any, K comparable](in <-chan T, keyFunc func(T) K, channels map[K]chan<- T) {
	defer func() {
		for _, ch := range channels {
			close(ch)
		}
	}()

	for t := range in {
		k := keyFunc(t)
		if ch, ok := channels[k]; ok {
			ch <- t
		}
	}
}
