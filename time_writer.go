package main

import (
	"bytes"
	"sync"
	"time"
)

type TimeWriter struct {
	Interval     time.Duration
	lastBounceAt time.Time
	timer        *time.Timer
	data         bytes.Buffer
	resChan      chan []byte
	stopChan     chan struct{}
	mutex        *sync.Mutex
}

func NewTimeWriter(duration time.Duration) *TimeWriter {
	b := &TimeWriter{
		Interval:     duration,
		lastBounceAt: time.Now(),
		timer:        time.NewTimer(duration),
		resChan:      make(chan []byte, 1),
		mutex:        new(sync.Mutex),
		stopChan:     make(chan struct{}, 1),
	}
	go b.start()
	return b
}

func (b *TimeWriter) start() {
LOOP:
	for {
		select {
		case tm := <-b.timer.C:
			if lb := b.lastBounceAt; tm.After(lb) && tm.Sub(lb) >= b.Interval {
				b.timer.Reset(b.Interval)
				b.resChan <- b.readAll()
			} else {
				b.timer.Reset(lb.Add(b.Interval).Sub(time.Now()))
			}
		case <-b.stopChan:
			break LOOP
		}
	}
	close(b.resChan)
}

func (b *TimeWriter) DataChan() <-chan []byte {
	return b.resChan
}

func (b *TimeWriter) Stop() {
	select {
	case b.stopChan <- struct{}{}:
		close(b.stopChan)
	default:
	}
}

func (b *TimeWriter) Write(data []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.lastBounceAt = time.Now()
	return b.data.Write(data)
}

func (b *TimeWriter) readAll() []byte {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.lastBounceAt = time.Now()
	data := b.data.Bytes()
	b.data.Reset()
	return data
}
