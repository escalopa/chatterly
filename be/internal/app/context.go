package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/escalopa/chatterly/internal/log"
)

func NewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		println() // print a newline after "^C" or "SIGTERM" to avoid log line concatenation
		log.Warn("received shutdown signal", log.String("signal", sig.String()))
		cancel()
	}()

	return ctx
}
