package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/mrturkmencom/gip/config"
	"github.com/mrturkmencom/gip/iptables"
	pb "github.com/mrturkmencom/gip/iptables/proto"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := config.ValidateConfigPath("/app/config.yml"); err != nil {
		panic(err)
	}
	conf, err := config.NewConfig("/app/config.yml")
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

	fmt.Printf("DockerService gRPC server is running at port %s...\n", gRPCPort)
	if err := gRPCEndpoint.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
