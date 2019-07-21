// Package logger exports a logger to print easily results obtained concurrently.
package logger

import (
	"io"
	"log"
	"sync"
)

// Logger is a simple type to print results obtained concurrently.
type Logger struct {
	ch chan string
	wg sync.WaitGroup
}

// New receives an io.Writer and a capacity to create a new Logger
// which will print the received messages on the corresponding io.Writer
// and whose channel will have the corresponding received capacity.
// At the end it returns a pointer to the new Logger.
func New(w io.Writer, cap int) *Logger {
	l := Logger{
		ch: make(chan string, cap),
	}

	log.SetOutput(w)

	l.wg.Add(1)
	go func() {
		for v := range l.ch {
			log.Println(v)
		}
		l.wg.Done()
	}()

	return &l
}

// Stop closes the *Logger's channel and waits until all the messages
// have been printed out.
func (l *Logger) Stop() {
	close(l.ch)
	l.wg.Wait()
}

// Println send a message (msg) to the *Logger's channel
// to be printed out.
func (l *Logger) Println(msg string) {
	l.ch <- msg
}
