package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/containerd/ttrpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	socket := flag.String("socket", "test.sock", "socket path")
	flag.Parse()

	s, err := ttrpc.NewServer(
		ttrpc.WithServerHandshaker(ttrpc.UnixSocketRequireSameUser()),
		ttrpc.WithUnaryServerInterceptor(unaryLogger),
		// https://github.com/containerd/ttrpc/issues/148
		// ttrpc.WithStreamServerInterceptor(streamLogger),
	)
	if err != nil {
		return err
	}
	defer s.Close()

	RegisterTestServiceService(s, &TestServer{})

	callback := func() {
		s.Shutdown(context.Background())
		log.Println("Server shutdown done, cleaning up..")
	}

	ctx, cancel := signalContextWithCallback(context.Background(), os.Interrupt, callback)
	defer cancel()

	conn, err := net.Listen("unix", *socket)
	if err != nil {
		return err
	}
	defer os.Remove(*socket)
	defer conn.Close()

	if err := s.Serve(ctx, conn); err != nil {
		return err
	}

	return nil
}

func signalContextWithCallback(ctx context.Context, sig os.Signal, callback func()) (context.Context, context.CancelFunc) {
	sigCtx, stop := signal.NotifyContext(ctx, sig)
	done := make(chan struct{}, 1)
	stopDone := make(chan struct{}, 1)

	go func() {
		defer func() { stopDone <- struct{}{} }()
		defer stop()
		select {
		case <-sigCtx.Done():
			fmt.Print("\r")
			log.Printf("Signal %s caught.\n", sig.String())
			log.Println("Press ctrl+c again to terminate the program immediately.")
			log.Println("Shutting down...")
			callback()
		case <-done:
		}
	}()

	cancelFunc := func() {
		done <- struct{}{}
		<-stopDone
	}

	return sigCtx, cancelFunc
}
