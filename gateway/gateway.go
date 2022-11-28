package main

import (
	"context"
	"flag"
	"fmt"
	"grpclb/balancer"
	registry "grpclb/etcd"
	gw "grpclb/protoc/gateway"
	"net/http"

	etcd "github.com/coreos/etcd/client"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	//服务器的地址
	endPoint = flag.String("echoEndPoint", "etcd:///", "grpc-server's address")
)

func run() error {

	etcdConfig := etcd.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}
	registry.RegisterResolver("etcd", etcdConfig, "/backend/services", "test", "1.0")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin))}
	err := gw.RegisterGreeterHandlerFromEndpoint(ctx, mux, *endPoint, opts)
	if err != nil {
		fmt.Printf("register faiiled:%v\n", err)
		return err
	}

	return http.ListenAndServe(":8086", mux)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
