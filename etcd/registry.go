package etcd

import (
	"google.golang.org/grpc/metadata"
)

type ServiceInfo struct {
	InstanceId string
	Name       string
	Version    string
	Address    string
	Metadata   metadata.MD
}

type Register interface {
	Register(service *ServiceInfo) error
	Unregister(service *ServiceInfo) error
	Close()
}
