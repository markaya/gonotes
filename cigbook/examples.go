package cigbook

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func SyncCondExample() {
	m := &sync.Mutex{}
	cond := sync.Cond{L: m}

	status := false

	go func() {
		cond.L.Lock()

		for !status {
			fmt.Println("I will execute this once and go to sleep")
			cond.Wait()
		}
		cond.L.Unlock()
	}()

	go func() {
		time.Sleep(1 * time.Second)

		cond.L.Lock()

		fmt.Println("Here we go")
		status = true

		cond.L.Unlock()

		cond.Signal()
		//cond.Broadcast()

	}()
}

// TODO: Test this generic version
func OrChannel[T any](channels ...<-chan T) <-chan T {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan T)
	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-OrChannel(append(channels[3:], orDone)...):
			}
		}
	}()

	return orDone
}

// NOTE: done channel is usually context.Done
func OrDone[T any](done <-chan struct{}, c <-chan T) <-chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valStream <- v:
				case <-done:
					return
				}
			}
		}
	}()
	return valStream
}

func TeeChannel[T any](done <-chan struct{}, in <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for val := range OrDone(done, in) {
			var o1, o2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case o1 <- val:
					o1 = nil
				case o2 <- val:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func TeeNWay[T any](done <-chan struct{}, in <-chan T, n int) []<-chan T {
	outs := make([]chan T, n)
	for i := range outs {
		outs[i] = make(chan T)
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()

		for val := range OrDone(done, in) {
			// Track which channels we've sent to
			sent := make([]bool, n)
			remaining := n

			for remaining > 0 {
				for i, ch := range outs {
					if sent[i] {
						continue
					}

					select {
					case <-done:
						return
					case ch <- val:
						sent[i] = true
						remaining--
					default:
						// Channel not ready, continue to next
					}
				}

				// If not all sent, yield to prevent busy-wait
				if remaining > 0 {
					runtime.Gosched()
				}
			}
		}
	}()

	readOnly := make([]<-chan T, n)
	for i, ch := range outs {
		readOnly[i] = ch
	}
	return readOnly
}

func TeeBuffered[T any](done <-chan struct{}, in <-chan T, bufferSize int) (<-chan T, <-chan T) {
	out1 := make(chan T, bufferSize)
	out2 := make(chan T, bufferSize)

	go func() {
		defer close(out1)
		defer close(out2)

		for val := range OrDone(done, in) {
			var o1, o2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case o1 <- val:
					o1 = nil
				case o2 <- val:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func Bridge[T any](done <-chan struct{}, chanStream <-chan <-chan T) <-chan T {
	valStream := make(chan T)

	go func() {
		defer close(valStream)

		for {
			var stream <-chan T
			select {
			case maybeStream, ok := <-chanStream:
				if !ok {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}

			for val := range OrDone(done, stream) {
				select {
				case valStream <- val:
				case <-done:
					return
				}
			}
		}
	}()

	return valStream
}
func BridgeBuffered[T any](done <-chan struct{}, chanStream <-chan <-chan T, bufferSize int) <-chan T {
	valStream := make(chan T, bufferSize)

	go func() {
		defer close(valStream)

		for {
			var stream <-chan T
			select {
			case maybeStream, ok := <-chanStream:
				if !ok {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}

			for val := range OrDone(done, stream) {
				select {
				case valStream <- val:
				case <-done:
					return
				}
			}
		}
	}()

	return valStream
}

func FanOut[T any, R any](
	done <-chan struct{},
	in <-chan T,
	workerFunc func(T) R,
	numWorkers int,
) []<-chan R {
	channels := make([]<-chan R, numWorkers)

	for i := 0; i < numWorkers; i++ {
		channels[i] = worker(done, in, workerFunc)
	}

	return channels
}

func worker[T any, R any](
	done <-chan struct{},
	in <-chan T,
	workerFunc func(T) R,
) <-chan R {
	out := make(chan R)

	go func() {
		defer close(out)
		for val := range OrDone(done, in) {
			select {
			case out <- workerFunc(val):
			case <-done:
				return
			}
		}
	}()

	return out
}

func FanIn[T any](done <-chan struct{}, channels ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan T) {
			defer wg.Done()
			for val := range OrDone(done, c) {
				select {
				case out <- val:
				case <-done:
					return
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
