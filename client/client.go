package main

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	registry "grpclb/etcd"
	"grpclb/interceptor"

	"grpclb/balancer"
	pb "grpclb/protoc/gateway"

	etcd "github.com/coreos/etcd/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	nCtx := metautils.NiceMD(md).ToOutgoing(ctx)
	return nCtx
}

func main() {

	etcdConfig := etcd.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}
	registry.RegisterResolver("etcd", etcdConfig, "/backend/services", "test", "1.0")
	//balancer.Register(&registry.etcdResolver{"etcd"})
	//b := grpc.RoundRobin(registry.etcdResolver{"etcd"})
	//grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig":[{"%s:{}"}]}`, balancer.RoundRobin))

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// //添加认证
	// //双向认证
	// clientPerm := "client.Pem"
	// clientKey := "client.key"
	// cert, err := tls.LoadX509KeyPair(clientPerm, clientKey)
	// if err != nil {
	// 	log.Fatalf("load keypair failed%v\n", err)
	// }
	//
	// //单向认证：客户端只对服务端的证书通过ca证书进行解析验证，不传递自己的证书
	// //双向认证：客户端要加载自己的证书，供服务端进行验证，通过x509.Load价值
	// //(1)单向认证：加载ca证书

	// //方式一
	// config, err := credentials.NewClientTLSFromFile("../cert/ca.crt", "*.fairzhang.com")
	// if err != nil {
	// 	fmt.Printf("new client tsl failed:%v", err)
	// 	return
	// }
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // 不用校验服务器证书
	}
	config := credentials.NewTLS(tlsConfig)

	//方式二
	// ca := "ca.crt"
	// caCert, err := ioutil.ReadFile(ca)
	// if err != nil {
	// 	log.Fatalf("load ca's file failed:%v\n", err)
	// }
	// certPool := x509.NewCertPool()
	// //(2)解析证书到certPool中
	// ok := certPool.AppendCertsFromPEM(caCert)
	// if !ok {
	// 	log.Fatalf("parse pem file failed\n")
	// }

	////单向认证
	//config := credentials.NewClientTLSFromCert(certPool, "")

	////双向认证
	//config := credentials.NewTLS(&tls.Config{
	//	Certificates: []tls.Certificate{cert},
	//	RootCAs:      certPool,
	//})

	//(3)添加连接时的认证选项
	//conn, err := grpc.DialContext(ctx, "etcd:///", grpc.WithInsecure(), grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin)), grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor1))
	//conn, err := grpc.DialContext(ctx, "etcd:///", grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin)), grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor1), grpc.WithTransportCredentials(config))
	//conn, err := grpc.DialContext(ctx, "etcd:///", grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin)), grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor1))
	//对于链式拦截器，调用顺序与传入的顺序相反，即此处先调用UnaryInterceptor2,再调用UnaryInterceptor1
	//conn, err := grpc.DialContext(ctx, "etcd:///", grpc.WithInsecure(), grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin)), grpc.WithChainUnaryInterceptor(interceptor.UnaryClientInterceptor1, interceptor.UnaryClientInterceptor2))

	//不使用域名解析和负载均衡，直接连接服务器
	conn, err := grpc.DialContext(ctx, "localhost:9006", grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, balancer.RoundRobin)), grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor1), grpc.WithTransportCredentials(config))
	if err != nil {
		panic(err)
	}
	client := pb.NewGreeterClient(conn)
	ticker := time.NewTicker(1 * time.Second)
	//ctx = ctxWithToken(context.Background(), "good_scheme1", "some_good_token")
	for t := range ticker.C {
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		} else {
			fmt.Printf("SayHello failed:%v\n", err)
		}
	}
}
