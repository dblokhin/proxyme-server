package main

import "sync"

// singleflight that forgets right after the main thread executed
type singleflight[K comparable, T any] struct {
	mu    sync.Mutex
	queue map[K]chan *singleflightResult[T]
}

type singleflightResult[T any] struct {
	v    T
	e    error
	done chan any
}

func (r *singleflightResult[T]) Result(value T, err error) {
	r.v, r.e = value, err
	close(r.done)
}

func (s *singleflight[K, T]) Do(key K, fn func() (T, error)) (T, error) {
	s.mu.Lock()

	if qu, ok := s.queue[key]; ok {
		s.mu.Unlock()
		res := &singleflightResult[T]{
			done: make(chan any),
		}
		qu <- res

		<-res.done

		select {
		case next := <-qu:
			next.Result(res.v, res.e)
		default:
		}

		return res.v, res.e
	}

	// make queue
	queue := make(chan *singleflightResult[T])
	s.queue[key] = queue
	s.mu.Unlock()

	// run the code
	res, err := fn()

	// forget key
	s.mu.Lock()
	delete(s.queue, key)
	s.mu.Unlock()

	// free queue
	select {
	case next := <-queue:
		next.Result(res, err)
	default:
	}

	return res, err
}
