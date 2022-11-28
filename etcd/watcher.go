package etcd

import (
	"context"
	"encoding/json"
	"sync"

	etcd_cli "github.com/coreos/etcd/client"
	"google.golang.org/grpc/resolver"
)

type Watcher struct {
	key     string
	keyapi  etcd_cli.KeysAPI
	watcher etcd_cli.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	addrs   []resolver.Address
}

func (w *Watcher) Close() {
	w.cancel()
	w.wg.Wait()
}

func newWatcher(key string, cli etcd_cli.Client) *Watcher {
	api := etcd_cli.NewKeysAPI(cli)
	watcher := api.Watcher(key, &etcd_cli.WatcherOptions{Recursive: true})
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		key:     key,
		keyapi:  api,
		watcher: watcher,
		ctx:     ctx,
		cancel:  cancel,
	}
	return w
}

func (w *Watcher) GetAllAddresses() []resolver.Address {
	resp, _ := w.keyapi.Get(w.ctx, w.key, &etcd_cli.GetOptions{Recursive: true})
	addrs := []resolver.Address{}
	for _, n := range resp.Node.Nodes {
		serviceInfo := ServiceInfo{}
		err := json.Unmarshal([]byte(n.Value), &serviceInfo)
		if err != nil {
			continue
		}
		addrs = append(addrs, resolver.Address{
			Addr:     serviceInfo.Address,
			Metadata: &serviceInfo.Metadata,
		})
	}
	return addrs
}

func (w *Watcher) Watch() chan []resolver.Address {
	out := make(chan []resolver.Address, 10)
	w.wg.Add(1)
	go func() {
		defer func() {
			close(out)
			w.wg.Done()
		}()
		w.addrs = w.GetAllAddresses()
		out <- w.addrs

		for {
			resp, err := w.watcher.Next(w.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
			}

			if resp.Node.Dir {
				continue
			}
			nodeData := ServiceInfo{}
			if resp.Action == "set" || resp.Action == "create" || resp.Action == "update" ||
				resp.Action == "delete" || resp.Action == "expire" {
				err := json.Unmarshal([]byte(resp.Node.Value), &nodeData)
				if err != nil {
					continue
				}
				addr := resolver.Address{Addr: nodeData.Address, Metadata: &nodeData.Metadata}
				changed := false
				switch resp.Action {
				case "set", "create":
					changed = w.addAddr(addr)
				case "update":
					changed = w.updateAddr(addr)
				case "delete", "expire":
					changed = w.removeAddr(addr)
				}
				if changed {
					out <- w.addrs
				}
			}
		}
	}()

	return out
}

func (w *Watcher) addAddr(addr resolver.Address) bool {
	for _, v := range w.addrs {
		if addr.Addr == v.Addr {
			return false
		}
	}
	w.addrs = append(w.addrs, addr)
	return true
}

func (w *Watcher) updateAddr(addr resolver.Address) bool {
	for i, v := range w.addrs {
		if addr.Addr == v.Addr {
			w.addrs[i] = addr
			return true
		}
	}
	return false
}

func (w *Watcher) removeAddr(addr resolver.Address) bool {
	for i, v := range w.addrs {
		if v.Addr == addr.Addr {
			w.addrs = append(w.addrs[:i], w.addrs[i+1:]...)
			return true
		}
	}
	return false
}

// // Close do nothing
// func (w *watcher) Close() {
// }
//
// // Next to return the updates
// func (w *watcher) Next() ([]*naming.Update, error) {
// 	// prefix is the etcd prefix/value to watch
// 	prefix := fmt.Sprintf("/%s/%s/", Prefix, w.re.serviceName)
// 	// check if is initialized
// 	if !w.isInitialized {
// 		// query addresses from etcd
// 		resp, err := w.client.Get(context.Background(), prefix, etcd3.WithPrefix())
// 		w.isInitialized = true
// 		if err == nil {
// 			addrs := extractAddrs(resp)
// 			//if not empty, return the updates or watcher new dir
// 			if l := len(addrs); l != 0 {
// 				updates := make([]*naming.Update, l)
// 				for i := range addrs {
// 					updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
// 				}
// 				return updates, nil
// 			}
// 		}
// 	}
// 	// generate etcd Watcher
// 	rch := w.client.Watch(context.Background(), prefix, etcd3.WithPrefix())
// 	for wresp := range rch {
// 		for _, ev := range wresp.Events {
// 			switch ev.Type {
// 			case mvccpb.PUT:
// 				return []*naming.Update{{Op: naming.Add, Addr: string(ev.Kv.Value)}}, nil
// 			case mvccpb.DELETE:
// 				return []*naming.Update{{Op: naming.Delete, Addr: string(ev.Kv.Value)}}, nil
// 			}
// 		}
// 	}
// 	return nil, nil
// }
// func extractAddrs(resp *etcd3.GetResponse) []string {
// 	addrs := []string{}
// 	if resp == nil || resp.Kvs == nil {
// 		return addrs
// 	}
// 	for i := range resp.Kvs {
// 		if v := resp.Kvs[i].Value; v != nil {
// 			addrs = append(addrs, string(v))
// 		}
// 	}
// 	return addrs
// }
