package main

import (
	"context"
	"os/signal"
	"syscall"

	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := log.NewLogger("", zapcore.DebugLevel)
	if err != nil {
		logger.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer logger.Close()

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		logger.Fatal("failed to bind to port 6379", zap.Error(err))
	}
	server := server.Server{
		Events:   make(chan server.Event),
		Listener: listener,
		Logger:   *logger,
		StoreData: make(map[string]any),
	}

	ctx, cancel := context.WithCancel(context.Background())

	go server.EventLoop(ctx)
	go server.ConnectionHandler(ctx)

	sigShutdown := make(chan os.Signal, 1)
	signal.Notify(sigShutdown, syscall.SIGTERM, syscall.SIGINT)

	<-sigShutdown
	cancel()
	logger.Info("server received shutdown signal")
}
