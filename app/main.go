package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	port := flag.Int("port", 6379, "specify the port that this reddis instance will listen on")

	var replicaof string
	flag.StringVar(&replicaof, "replicaof", "", "specify the hostname and port that this instance should be a replica of")
	flag.Parse()

	// This flag may be formatted as "hostname port" so we need to turn this into an actual address
	replicaof = strings.ReplaceAll(replicaof, " ", ":")

	logger, err := log.NewLogger("", zapcore.InfoLevel)
	if err != nil {
		logger.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer logger.Close()

	ctx, cancel := context.WithCancel(context.Background())

	serverOpts := server.ServerOptions{
		Port: port,
	}

	logger.AddMetadata(zap.Int("serverListenPort", *port))

	if replicaof == "" {
		logger.AddMetadata(zap.String("nodeType", string(server.MasterNodeType)))
		server, err := server.NewMasterServer(*logger, serverOpts)
		if err != nil {
			logger.Fatal("failed to initialize master server", zap.Error(err))
		}

		err = server.Run(ctx)
		if err != nil {
			logger.Fatal("failed to run master server", zap.Error(err))
		}
	} else {
		logger.AddMetadata(zap.String("nodeType", string(server.ReplicaNodeType)))
		server, err := server.NewReplicaServer(*logger, replicaof, serverOpts)
		if err != nil {
			logger.Fatal("failed to initialize replica server", zap.Error(err))
		}
		err = server.Run(ctx)
		if err != nil {
			logger.Fatal("failed to run replica server", zap.Error(err))
		}
	}

	sigShutdown := make(chan os.Signal, 1)
	signal.Notify(sigShutdown, syscall.SIGTERM, syscall.SIGINT)

	<-sigShutdown
	cancel()
	logger.Info("server received shutdown signal")
}
