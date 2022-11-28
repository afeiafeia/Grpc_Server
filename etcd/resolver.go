package etcd

import (
	//etcd3 "github.com/coreos/etcd/clientv3"

	"sync"

	etcd_cli "github.com/coreos/etcd/client"
	"google.golang.org/grpc/resolver"
)

// // resolver is the implementaion of grpc.naming.Resolver
// type resolver struct {
// 	serviceName string // service name to resolve
// }
//
// // NewResolver return resolver with service name
// func NewResolver(serviceName string) *resolver {
// 	return &resolver{serviceName: serviceName}
// }
//
// // Resolve to resolve the service from etcd, target is the dial address of etcd
// // target example: "http://127.0.0.1:2379,http://127.0.0.1:12379,http://127.0.0.1:22379"
// func (re *resolver) Resolve(target string) (naming.Watcher, error) {
// 	if re.serviceName == "" {
// 		return nil, errors.New("grpclb: no service name provided")
// 	}
// 	// generate etcd client
// 	client, err := etcd3.New(etcd3.Config{
// 		Endpoints: strings.Split(target, ","),
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("grpclb: creat etcd3 client failed: %s", err.Error())
// 	}
// 	// Return watcher
// 	return &watcher{re: re, client: *client}, nil
// }

type etcdResolver struct {
	scheme        string
	etcdConfig    etcd_cli.Config
	etcdWatchPath string
	watcher       *Watcher
	target        resolver.Target
	cc            resolver.ClientConn
	wg            sync.WaitGroup
}

// build创建一个解析器resolver,并执行它的start函数开始监视注册中心的信息
// BUild方法是Builder接口的方法之一，该接口有两个方法，另一个是Scheme,用于返回resolver的scheme值
func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	etcdCli, err := etcd_cli.New(r.etcdConfig)
	if err != nil {
		return nil, err
	}

	r.target = target
	r.cc = cc
	r.watcher = newWatcher(r.etcdWatchPath, etcdCli)
	r.start()
	return r, nil
}

func (r *etcdResolver) Scheme() string {
	return r.scheme
}

// start函数开始监视
func (r *etcdResolver) start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		out := r.watcher.Watch()
		for addr := range out {
			//fmt.Println(addr)
			r.cc.UpdateState(resolver.State{Addresses: addr})
		}
	}()
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {

}
func (r etcdResolver) Close() {
	r.watcher.Close()
	r.wg.Wait()
}

func RegisterResolver(scheme string, etcdConfig etcd_cli.Config, registryDir, srvName, srvVersion string) {
	resolver.Register(&etcdResolver{
		scheme:        scheme,
		etcdConfig:    etcdConfig,
		etcdWatchPath: registryDir + "/" + srvName + "/" + srvVersion,
	})
}
