package grpctest

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/containerd/ttrpc"
)

func ScriptMain() int {
	if err := Run(); err != nil {
		return 1
	}
	return 0
}

func Run() error {
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
		if ctx.Err() == context.Canceled && errors.Is(err, ttrpc.ErrServerClosed) {
			return nil
		}
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
