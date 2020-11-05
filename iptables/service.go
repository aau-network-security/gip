package iptables

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/mrturkmencom/gip/config"
	pb "github.com/mrturkmencom/gip/iptables/proto"
	authLib "github.com/mrturkmencom/go-helpers/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type iptablesservice struct {
	auth   authLib.Authenticator
	config *config.Config
}

func InitializeIPTService(conf *config.Config) *iptablesservice {
	iptService := &iptablesservice{
		auth:   authLib.NewAuthenticator(conf.IPTService.AUTH.SignKey, conf.IPTService.AUTH.AuthKey),
		config: conf,
	}
	return iptService
}

func (ipt *iptablesservice) CreateAcceptRule(ctx context.Context, req *pb.AcceptRequest) (*pb.AcceptReply, error) {
	in := req.Input
	out := req.Output
	srv := IPTables{}
	if err := srv.SetAcceptRule(in, out); err != nil {
		return nil, err
	}
	return &pb.AcceptReply{}, nil
}

func (ipt *iptablesservice) CreateAcceptWithState(ctx context.Context, req *pb.AcceptRequest) (*pb.AcceptReply, error) {
	in := req.Input
	out := req.Output
	srv := IPTables{}
	if err := srv.CheckWhoCreatesConn(in, out); err != nil {
		return nil, err
	}
	return &pb.AcceptReply{}, nil
}
func (ipt *iptablesservice) DropForward(ctx context.Context, req *pb.FlushRequest) (*pb.Respond, error) {
	srv := IPTables{}

	if err := srv.DropExistingRule(Chain(req.Chain)); err != nil {
		return nil, err
	}
	return &pb.Respond{}, nil
}

func GetCreds(conf config.Config) (credentials.TransportCredentials, error) {
	log.Printf("Preparing credentials for RPC")

	certificate, err := tls.LoadX509KeyPair(conf.IPTService.TLS.CertFile, conf.IPTService.TLS.CertKey)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authorityS
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(conf.IPTService.TLS.CAFile)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}
	// CA file for let's encrypt is located under domain conf as `chain.pem`
	// pass chain.pem location
	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append client certs")
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	return creds, nil
}

// SecureConn enables communication over secure channel
func SecureConn(conf *config.Config) ([]grpc.ServerOption, error) {
	if conf.IPTService.TLS.Enabled {
		log.Info().Msgf("Conf cert-file: %s, cert-key: %s ca: %s", conf.IPTService.TLS.CertFile, conf.IPTService.TLS.CertKey, conf.IPTService.TLS.CAFile)
		creds, err := GetCreds(*conf)

		if err != nil {
			return []grpc.ServerOption{}, errors.New("Error on retrieving certificates: " + err.Error())
		}
		log.Printf("Server is running in secure mode !")
		return []grpc.ServerOption{grpc.Creds(creds)}, nil
	}
	return []grpc.ServerOption{}, nil
}

// AddAuth adds authentication to gRPC server
func (d *iptablesservice) AddAuth(opts ...grpc.ServerOption) *grpc.Server {
	streamInterceptor := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := d.auth.AuthenticateContext(stream.Context()); err != nil {
			return err
		}
		return handler(srv, stream)
	}

	unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := d.auth.AuthenticateContext(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}

	opts = append([]grpc.ServerOption{
		grpc.StreamInterceptor(streamInterceptor),
		grpc.UnaryInterceptor(unaryInterceptor),
	}, opts...)
	return grpc.NewServer(opts...)
}
