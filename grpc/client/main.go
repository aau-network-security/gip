package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
)

type Creds struct {
	Token    string
	Insecure bool
}

func (c Creds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"token": string(c.Token),
	}, nil
}

func (c Creds) RequireTransportSecurity() bool {
	return !c.Insecure
}

func main() {
	var conn *grpc.ClientConn
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iptservice": "test",
	})

	tokenString, err := token.SignedString([]byte("test"))
	if err != nil {
		fmt.Println("Error creating the token")
	}

	authCreds := Creds{Token: tokenString}
	dialOpts := []grpc.DialOption{}
	authCreds.Insecure = true
	dialOpts = append(dialOpts,
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(authCreds))

	conn, err = grpc.Dial(":4444", dialOpts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	//ctx := context.Background()
	//cli := pb.NewIPTablesClient(conn)

	//fmt.Printf("Close Container Response %s \n", closeResp.Msg)
}
