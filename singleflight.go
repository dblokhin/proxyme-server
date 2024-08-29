package main

import "sync"

// singleflight that forgets right after the main thread executed
type singleflight[K comparable, T any] struct {
	mu sync.Mutex
	m  map[K]*singleflightResult[T]
}

type singleflightResult[T any] struct {
	v    T
	e    error
	done chan any
}

func (s *singleflight[K, T]) Do(key K, fn func() (T, error)) (T, error) {
	s.mu.Lock()

	if result, ok := s.m[key]; ok {
		s.mu.Unlock()
		<-result.done

		return result.v, result.e
	}

	result := &singleflightResult[T]{
		done: make(chan any),
	}
	s.m[key] = result
	s.mu.Unlock()

	// run the code
	res, err := fn()
	result.v, result.e = res, err

	// forget key
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()

	// let waiters go
	close(result.done)

	return res, err
}
