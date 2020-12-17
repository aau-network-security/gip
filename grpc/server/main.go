package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/mrturkmencom/gip/config"
	"github.com/mrturkmencom/gip/iptables"
	pb "github.com/mrturkmencom/gip/iptables/proto"
	"google.golang.org/grpc/reflection"
)

var (
	CONFIG_FILE = os.Getenv("CONFIG_FILE")
)

func main() {
	if err := config.ValidateConfigPath(CONFIG_FILE); err != nil {
		panic(err)
	}
	conf, err := config.NewConfig(CONFIG_FILE)
	if err != nil {
		panic(err)
	}
	gRPCPort := strconv.FormatUint(uint64(conf.IPTService.Domain.Port), 10)

	lis, err := net.Listen("tcp", ":"+gRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ipTService := iptables.InitializeIPTService(conf)
	opts, err := iptables.SecureConn(conf)
	if err != nil {
		log.Fatalf("failed to retrieve secure options %s", err.Error())
	}
	gRPCEndpoint := ipTService.AddAuth(opts...)
	reflection.Register(gRPCEndpoint)
	pb.RegisterIPTablesServer(gRPCEndpoint, ipTService)

	fmt.Printf("gip: gRPC service is running at port %s...\n", gRPCPort)
	if err := gRPCEndpoint.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
