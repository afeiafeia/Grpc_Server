package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	registry "grpclb/etcd"
	"grpclb/interceptor"

	//pb "grpclb/protoc"
	pb "grpclb/protoc/gateway"

	"grpclb/common"

	etcd_cli "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	nodeId = flag.String("node", "ndoe2", "node_ID")
	port   = flag.Int("port", 9876, "listening port")
	reg    = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

var (
	commonAuthToken = "some_good_token"
)

type RpcServer struct {
	addr string
	s    *grpc.Server
	pb.UnimplementedGreeterServer
}

func NewRpcServer(addr string) *RpcServer {
	//单一拦截器
	//s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor))

	// //链式拦截器
	// scheme := "good_scheme"
	// token := commonAuthToken
	// authFun := auth.BuildDummyAuthFunction(scheme, token)

	// //添加安全传输:单向认证
	// //单向认证：客户端对服务端进行认证，服务端不对客户端进行认证，一般即是使用此种单向认证
	// //双向认证：服务端也对客户端的身份进行验证，一般是用于企业内部，限制特定客户端的访问
	// permFile := "../cert/server.pem"
	// keyFile := "../cert/server.key"
	// creds, err := credentials.NewServerTLSFromFile(permFile, keyFile)
	// if err != nil {
	// 	log.Fatalf("create tls form file failed:%v\n", err)
	// 	return nil
	// }

	// //双向认证
	// //(1)加载服务端证书和私钥
	// cert, err := tls.LoadX509KeyPair(permFile, keyFile)
	// if err != nil {
	// 	log.Fatalf("load keypair failed:%v\n", err)
	// 	return nil
	// }
	// //(2)加载根证书(ca)
	// //创建空certPool
	// certPool := x509.NewCertPool()
	// caFile := "ca.crt"
	// //读取整改书
	// ca, err := ioutil.ReadFile(caFile)
	// if err != nil {
	// 	log.Fatalf("read ca's file failed%v\n", err)
	// 	return nil
	// }
	// //解析所传入的编码证书，解析成功后会存入certPool
	// certPool.AppendCertsFromPEM(ca)
	// //构造tls所使用的TransportCredential
	// creds := credentials.NewTLS(&tls.Config{
	// 	Certificates: []tls.Certificate{cert},
	// 	ClientAuth:   tls.RequireAndVerifyClientCert,
	// 	ClientCAs:    certPool,
	// })

	//链式拦截器从后往前调用
	//s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.UnaryServerInterceptor, auth.UnaryServerInterceptor(authFun)), grpc.Creds(creds))
	// s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.UnaryServerInterceptor, auth.UnaryServerInterceptor(authFun)))
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.UnaryServerInterceptor))
	rs := &RpcServer{
		addr: addr,
		s:    s,
	}
	return rs
}

func (s *RpcServer) Run() {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Printf("failed to listen:%v", err)
		return
	}
	log.Printf("rpc listening on::%s", s.addr)
	pb.RegisterGreeterServer(s.s, s)
	s.s.Serve(listener)
}
func (s *RpcServer) Stop() {
	s.s.GracefulStop()
}

// SayHello implements helloworld.GreeterServer
func (s *RpcServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name + "this is " + *nodeId}, nil
}

func StartService() {
	etcdConfig := etcd_cli.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}

	service := &registry.ServiceInfo{
		InstanceId: *nodeId,
		Name:       "test",
		Version:    "1.0",
		Address:    fmt.Sprintf("127.0.0.1:%d", *port),
		Metadata:   metadata.Pairs(common.WeightKey, "1"),
	}

	//创建etcd客户端并记录注册的服务器信息
	etcdRegistry, err := registry.NewRegistrar(
		&registry.Config{
			EtcdConfig:  etcdConfig,
			RegistryDir: "/backend/services",
			Ttl:         10 * time.Second,
		})
	if err != nil {
		log.Printf("register failed:%v%", err)
		return
	}

	server := NewRpcServer(fmt.Sprintf("0.0.0.0:%d", *port))
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		server.Run()
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		etcdRegistry.Register(service)
		wg.Done()
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	etcdRegistry.Unregister(service)
	server.Stop()
	wg.Wait()

}

func main() {
	flag.Parse()
	StartService()
}
