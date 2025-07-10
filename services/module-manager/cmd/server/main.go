package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/drs-api/services/module-manager/internal/config"
	"github.com/drs-api/services/module-manager/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port       = flag.Int("port", 0, "The server port (0 = use config file)")
	configFile = flag.String("config", "", "Path to config file")
)

func main() {
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error
	
	if *configFile != "" {
		cfg, err = config.LoadConfig(*configFile)
	} else {
		cfg, err = config.LoadConfigFromDefault()
	}
	
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Override port from command line if specified
	serverPort := cfg.Server.Port
	if *port != 0 {
		serverPort = *port
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	service.RegisterServices(s, cfg)
	
	reflection.Register(s)

	log.Printf("Starting gRPC server on port %d", serverPort)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	s.GracefulStop()
}