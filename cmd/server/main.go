package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/laneblog/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server stopped with error: %v", err)
		}
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := application.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown failed: %v", err)
		}
	}
}
