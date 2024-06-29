package main

import (
	"context"
	"flag"
	"os/signal"
	"strings"
	"syscall"

	"os"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	port := flag.Int("port", 6379, "specify the port that this reddis instance will listen on")

	var replicaof string
	flag.StringVar(&replicaof, "replicaof", "", "specify the hostname and port that this instance should be a replica of")
	// This flag may be formatted as "hostname port" so we need to turn this into an actual address
	replicaof = strings.ReplaceAll(replicaof, " ", ":")

	flag.Parse()

	logger, err := log.NewLogger("", zapcore.InfoLevel)
	if err != nil {
		logger.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer logger.Close()

	server, err := server.NewServer(*logger, &server.ServerOptions{
		Port:      port,
		ReplicaOf: &replicaof,
	})
	if err != nil {
		logger.Fatal("failed to initialize server", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go server.EventLoop(ctx)
	go server.ConnectionHandler(ctx)
	go server.ExpiryLoop(ctx)

	sigShutdown := make(chan os.Signal, 1)
	signal.Notify(sigShutdown, syscall.SIGTERM, syscall.SIGINT)

	<-sigShutdown
	cancel()
	logger.Info("server received shutdown signal")
}
